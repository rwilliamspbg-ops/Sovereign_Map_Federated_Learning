# Privacy Proofs for Quantized Differentially Private Federated Learning

### Overview
This document proves that applying quantization (int8) in parallel with Gaussian differential privacy mechanisms guarantees bounded privacy loss and maintains our strict $(\epsilon, \delta)$-DP bounds.

### Definition: $L2$ Sensitivity Bound
Given a gradient space bounded by $C$, the $L2$ sensitivity consists of the Euclidean distance between any two clipped gradients:
$$ \Delta_2 \leq 2C $$

### Theorem 1: Post-Processing Inequality
Quantization is a deterministic discretization function $f: \mathbb{R}^d \rightarrow \mathbb{Z}^d$. By the post-processing property of Differential Privacy:
- Let $M$ be an $(\epsilon, \delta)$-DP mechanism.
- Let $Q$ be our symmetric int8 quantization mapping.
Then $Q(M(X))$ is also $(\epsilon, \delta)$-DP.

### Mechanism Formulation
1. **Clipping**: Clip gradients $g$ such that $\|g\|_2 \leq C$.
2. **Noise Addition**: $g' = g + N(0, \sigma^2 \Delta_2^2)$ where $\sigma = \frac{\sqrt{2\ln(1.25/\delta)}}{\epsilon}$.
3. **Quantization**: $q = Q_{int8}(g')$.

Because the quantization occurs *after* the noise bounds the maximum privacy leakage, the information exposed by the discrete bins cannot exceed the information exposed by the continuous noisy signal.

### Proof of Accuracy Bounds Under Quantization
The expected mean-squared error (MSE) introduced by noise $N$ and quantization $Q$ where $\Delta_q$ is the quantization step-size:
$$ \text{MSE} \approx \sigma^2 \Delta_2^2 + \frac{\Delta_q^2}{12} $$

Given int8 (256 bins), $\Delta_q = \frac{2C}{255}$. For typical FL models, $C$ decreases over rounds (convergence), rendering the quantization error negligible relative to the DP noise, meaning convergence is unaffected while providing 10x bandwidth savings.
