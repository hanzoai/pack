package resolver

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/mise"
)

const (
	DefaultSource = "railpack default"
)

type Resolver struct {
	mise     *mise.Mise
	packages map[string]*RequestedPackage
}

type RequestedPackage struct {
	Name    string
	Version string
	Source  string
}

type ResolvedPackage struct {
	Name             string  `json:"name"`
	RequestedVersion *string `json:"requestedVersion,omitempty"`
	ResolvedVersion  *string `json:"resolvedVersion,omitempty"`
	Source           string  `json:"source"`
}

type PackageRef struct {
	Name string
}

func NewRequestedPackage(name, defaultVersion string) *RequestedPackage {
	return &RequestedPackage{
		Name:    name,
		Version: defaultVersion,
		Source:  DefaultSource,
	}
}

func (p *RequestedPackage) SetVersion(version, source string) *RequestedPackage {
	p.Version = version
	p.Source = source
	return p
}

func NewResolver(miseDir string) (*Resolver, error) {
	mise, err := mise.New(miseDir)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		mise:     mise,
		packages: make(map[string]*RequestedPackage),
	}, nil
}

func (r *Resolver) ResolvePackages() (map[string]*ResolvedPackage, error) {
	resolvedPackages := make(map[string]*ResolvedPackage)

	for name, pkg := range r.packages {
		fuzzyVersion := resolveToFuzzyVersion(pkg.Version)
		latestVersion, err := r.mise.GetLatestVersion(name, fuzzyVersion)

		if err != nil {
			return nil, err
		}

		log.Debugf("Resolved package version %s %s to %s from %s", name, pkg.Version, latestVersion, pkg.Source)

		resolvedPkg := &ResolvedPackage{
			Name:             name,
			RequestedVersion: &pkg.Version,
			ResolvedVersion:  &latestVersion,
			Source:           pkg.Source,
		}

		resolvedPackages[name] = resolvedPkg
	}

	return resolvedPackages, nil
}

func (r *Resolver) Get(name string) *RequestedPackage {
	return r.packages[name]
}

func (r *Resolver) Default(name, defaultVersion string) PackageRef {
	r.packages[name] = NewRequestedPackage(name, defaultVersion)
	return PackageRef{Name: name}
}

func (r *Resolver) Version(ref PackageRef, version, source string) PackageRef {
	if pkg, exists := r.packages[ref.Name]; exists {
		pkg.SetVersion(strings.TrimSpace(version), source)
	}
	return ref
}
