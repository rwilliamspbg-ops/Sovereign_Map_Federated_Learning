from pathlib import Path

req_path = Path("requirements-backend.txt")
content = req_path.read_text()

# Fix asyncio version
content = content.replace("asyncio==4.0.0", "asyncio") # asyncio is part of stdlib, but some packages might depend on it. In modern python it should not be pinned to a non-existent version.

path = Path("requirements-backend.txt")
path.write_text(content)
print("Requirements fixed.")
