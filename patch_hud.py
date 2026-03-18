import re

with open("frontend/src/HUD.jsx", "r") as f:
    content = f.read()

# Add a visual equalizer
visualizer = """
          <div className="active-visualizer" style={{ marginTop: '10px' }}>
            <div className="bar"></div>
            <div className="bar"></div>
            <div className="bar"></div>
            <div className="bar"></div>
            <div className="bar"></div>
            <div style={{marginLeft: '10px', fontSize: '0.8rem', color: 'var(--brand)'}}>NEURAL LINK ACTIVE</div>
          </div>
"""

content = content.replace(
    "<h2>🎤 Operations Connect</h2>", "<h2>🎤 Operations Connect</h2>" + visualizer
)
content = content.replace(
    ">Trigger FL Round<", ' className="btn-primary">[ INTIALIZE FL SEQUENCE ]<'
)
content = content.replace(
    "Synchronizing Node State...", "ESTABLISHING SECURE NEURAL LINK..."
)

with open("frontend/src/HUD.jsx", "w") as f:
    f.write(content)
