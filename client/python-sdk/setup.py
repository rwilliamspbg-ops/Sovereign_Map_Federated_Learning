from setuptools import setup, find_packages

setup(
    name="sovereign-map-sdk",
    version="0.1.0",
    description="Python SDK for interacting with Sovereign Map node runtime and model distribution APIs",
    packages=find_packages(),
    python_requires=">=3.9",
    install_requires=[
        "requests>=2.28.0",
        "numpy>=1.24.0",
    ],
)
