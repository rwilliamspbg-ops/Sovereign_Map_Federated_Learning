import re

with open("frontend/src/App.jsx", "r") as f:
    content = f.read()


def repl(m):
    return """  const submitVoiceQuery = async () => {
    if (!voiceQuery.trim()) {
      return;
    }

    try {
      const response = await fetch(`${API_BASE}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: voiceQuery })
      });
      if (response.ok) {
        const result = await response.json();
        setVoiceResponse(result.response);
      } else {
        setVoiceResponse('Unable to retrieve system status');
      }
    } catch (err) {
      console.error('Voice query error:', err);
      setVoiceResponse('Connection error communicating with node');
    }
  };"""


content = re.sub(
    r"const submitVoiceQuery.*?catch \(err\) \{[^}]+\}\s+};",
    repl,
    content,
    flags=re.DOTALL,
)

with open("frontend/src/App.jsx", "w") as f:
    f.write(content)
