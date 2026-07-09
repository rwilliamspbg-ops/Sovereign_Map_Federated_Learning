with open("frontend/src/App.jsx", "r") as f:
    text = f.read()

# Change TRUST_API_BASE to just be API_BASE
import re

text = re.sub(r"const TRUST_API_BASE =[^;]+;", "const TRUST_API_BASE = API_BASE;", text)

with open("frontend/src/App.jsx", "w") as f:
    f.write(text)
