import numpy as np
import pandas as pd
import os

print("PATH:", os.environ['PATH'])
path = os.environ['PATH']
assert path.startswith("/app/.venv/bin"), f"Expected PATH to start with /app/.venv/bin but got {path}"


print("numpy", np.__version__)
print("pandas", pd.__version__)

print("Hello from UV")
