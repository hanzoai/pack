// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: "Railpack Docs",
      social: {
        github: "https://github.com/railwayapp/railpack",
      },
      sidebar: [
        {
          label: "Start Here",
          items: [{ label: "Getting Started", link: "/getting-started" }],
        },
        {
          label: "Guides",
          items: [
            {
              label: "Building with CLI and BuildKit",
              link: "/guides/building-with-cli",
            },
            {
              label: "Building with a Custom Frontend",
              link: "/guides/custom-frontend",
            },
          ],
        },
        {
          label: "Configuration",
          items: [
            { label: "Configuration File", link: "/config/file" },
            {
              label: "Environment Variables",
              link: "/config/environment-variables",
            },
          ],
        },
        {
          label: "Languages",
          items: [
            { label: "Node.js", link: "/languages/node" },
            { label: "Python", link: "/languages/python" },
            { label: "Go", link: "/languages/golang" },
            { label: "PHP", link: "/languages/php" },
          ],
        },
        {
          label: "CLI Reference",
          items: [{ label: "CLI Commands", link: "/cli-reference" }],
        },
        {
          label: "Architecture",
          items: [
            { label: "High Level Overview", link: "/architecture/overview" },
            {
              label: "Package Resolution",
              link: "/architecture/package-resolution",
            },
            { label: "Plan Generation", link: "/architecture/plan-generation" },
            { label: "Secrets and Environment", link: "/architecture/secrets" },
            {
              label: "Previous Versions",
              link: "/architecture/previous-versions",
            },
            { label: "BuildKit Generation", link: "/architecture/buildkit" },
            { label: "Caching", link: "/architecture/caching" },
          ],
        },
      ],
    }),
  ],
});
