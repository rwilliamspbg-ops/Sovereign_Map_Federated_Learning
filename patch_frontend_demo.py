with open("frontend/src/BrowserFLDemo.jsx", "r") as f:
    text = f.read()

import re

text = re.sub(
    r"const TRAINING_API_BASE =[^;]+;", "const TRAINING_API_BASE = API_BASE;", text
)

with open("frontend/src/BrowserFLDemo.jsx", "w") as f:
    f.write(text)
