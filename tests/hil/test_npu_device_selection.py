import importlib.util
import sys
import types
from pathlib import Path


class _FakeNPU:
    def __init__(self, available: bool = False, raise_on_check: bool = False, device_count: int = 1):
        self._available = available
        self._raise_on_check = raise_on_check
        self._device_count = device_count
        self.selected_device = None
        self._memory_per_device = {i: 80 * 1024 * 1024 * 1024 for i in range(device_count)}  # 80GB each

    def is_available(self) -> bool:
        if self._raise_on_check:
            raise RuntimeError("NPU runtime unavailable")
        return self._available

    def set_device(self, device):
        self.selected_device = device

    def device_count(self) -> int:
        """Return number of available NPU devices"""
        return self._device_count if self._available else 0

    def get_device_properties(self, device_id: int) -> types.SimpleNamespace:
        """Return device properties including memory"""
        if not self._available or device_id >= self._device_count:
            raise RuntimeError(f"NPU device {device_id} not available")
        return types.SimpleNamespace(
            name=f"Huawei Ascend NPU {device_id}",
            total_memory=self._memory_per_device.get(device_id, 80 * 1024 * 1024 * 1024),
            capability=(8, 0),
        )

    def empty_cache(self):
        """Clear NPU memory cache"""
        pass


class _FakeCUDA:
    def __init__(self, available: bool = False, device_count: int = 1, device_prefix: str = "NVIDIA Tesla V"):
        self._available = available
        self._device_count = device_count
        self._device_prefix = device_prefix
        self._memory_per_device = {i: 24 * 1024 * 1024 * 1024 for i in range(device_count)}  # 24GB each

    def is_available(self) -> bool:
        return self._available

    def device_count(self) -> int:
        """Return number of available CUDA devices"""
        return self._device_count if self._available else 0

    def get_device_properties(self, device_id: int) -> types.SimpleNamespace:
        """Return device properties including memory"""
        if not self._available or device_id >= self._device_count:
            raise RuntimeError(f"CUDA device {device_id} not available")
        return types.SimpleNamespace(
            name=f"{self._device_prefix}{100 + device_id}",
            total_memory=self._memory_per_device.get(device_id, 24 * 1024 * 1024 * 1024),
            capability=(7, 0),
        )

    def empty_cache(self):
        """Clear CUDA memory cache"""
        pass

    def get_device_name(self, device_id: int) -> str:
        """Return device name"""
        if not self._available or device_id >= self._device_count:
            raise RuntimeError(f"CUDA device {device_id} not available")
        return f"{self._device_prefix}{100 + device_id}"


class _FakeXPU:
    def __init__(self, available: bool = False, device_count: int = 1):
        self._available = available
        self._device_count = device_count
        self.selected_device = None
        self._memory_per_device = {i: 16 * 1024 * 1024 * 1024 for i in range(device_count)}

    def is_available(self) -> bool:
        return self._available

    def set_device(self, device):
        self.selected_device = device

    def device_count(self) -> int:
        return self._device_count if self._available else 0

    def get_device_properties(self, device_id: int) -> types.SimpleNamespace:
        if not self._available or device_id >= self._device_count:
            raise RuntimeError(f"XPU device {device_id} not available")
        return types.SimpleNamespace(
            name=f"Intel XPU Flex {device_id}",
            total_memory=self._memory_per_device.get(device_id, 16 * 1024 * 1024 * 1024),
            capability=(1, 0),
        )

    def synchronize(self):
        pass

    def empty_cache(self):
        pass

    def get_device_name(self, device_id: int) -> str:
        if not self._available or device_id >= self._device_count:
            raise RuntimeError(f"XPU device {device_id} not available")
        return f"Intel XPU Flex {device_id}"


class _FakeMPS:
    def __init__(self, available: bool = False):
        self._available = available

    def is_available(self) -> bool:
        return self._available


def _install_module_stubs(
    monkeypatch,
    *,
    npu_available=False,
    npu_raises=False,
    npu_device_count=1,
    cuda_available=False,
    cuda_device_count=1,
    cuda_backend="cuda",
    xpu_available=False,
    xpu_device_count=1,
    mps_available=False,
):
    fake_torch = types.ModuleType("torch")
    fake_torch.npu = _FakeNPU(available=npu_available, raise_on_check=npu_raises, device_count=npu_device_count)
    fake_torch.cuda = _FakeCUDA(
        available=cuda_available,
        device_count=cuda_device_count,
        device_prefix=("AMD Radeon Instinct MI" if cuda_backend == "rocm" else "NVIDIA Tesla V"),
    )
    fake_torch.xpu = _FakeXPU(available=xpu_available, device_count=xpu_device_count)
    fake_torch.device = lambda name: name
    fake_torch.__version__ = "2.1.0"
    fake_torch.version = types.SimpleNamespace(cuda="12.1", hip=("6.1" if cuda_backend == "rocm" else None))
    fake_torch.backends = types.SimpleNamespace(mps=_FakeMPS(available=mps_available))

    fake_torch_nn = types.ModuleType("torch.nn")
    fake_torch_nn.Module = type("Module", (), {})
    fake_torch_nn_functional = types.ModuleType("torch.nn.functional")

    fake_torch_utils = types.ModuleType("torch.utils")
    fake_torch_utils_data = types.ModuleType("torch.utils.data")
    fake_torch_utils_data.DataLoader = type("DataLoader", (), {})
    fake_torch_utils_data.Subset = type("Subset", (), {})

    fake_flwr = types.ModuleType("flwr")
    fake_flwr.client = types.SimpleNamespace(NumPyClient=type("NumPyClient", (), {}))

    fake_opacus = types.ModuleType("opacus")
    fake_opacus.PrivacyEngine = type("PrivacyEngine", (), {})

    fake_torchvision = types.ModuleType("torchvision")
    fake_torchvision.datasets = types.SimpleNamespace(MNIST=type("MNIST", (), {}))
    fake_torchvision.transforms = types.SimpleNamespace(
        Compose=lambda *args, **kwargs: None
    )

    fake_numpy = types.ModuleType("numpy")
    fake_numpy.ndarray = object
    fake_numpy.mean = lambda values: sum(values) / len(values) if values else 0.0

    monkeypatch.setitem(sys.modules, "torch", fake_torch)
    monkeypatch.setitem(sys.modules, "torch.nn", fake_torch_nn)
    monkeypatch.setitem(sys.modules, "torch.nn.functional", fake_torch_nn_functional)
    monkeypatch.setitem(sys.modules, "torch.utils", fake_torch_utils)
    monkeypatch.setitem(sys.modules, "torch.utils.data", fake_torch_utils_data)
    monkeypatch.setitem(sys.modules, "flwr", fake_flwr)
    monkeypatch.setitem(sys.modules, "opacus", fake_opacus)
    monkeypatch.setitem(sys.modules, "torchvision", fake_torchvision)
    monkeypatch.setitem(sys.modules, "numpy", fake_numpy)

    return fake_torch


def _load_client_module(
    monkeypatch,
    *,
    npu_available=False,
    npu_raises=False,
    npu_device_count=1,
    cuda_available=False,
    cuda_device_count=1,
    cuda_backend="cuda",
    xpu_available=False,
    xpu_device_count=1,
    mps_available=False,
):
    fake_torch = _install_module_stubs(
        monkeypatch,
        npu_available=npu_available,
        npu_raises=npu_raises,
        npu_device_count=npu_device_count,
        cuda_available=cuda_available,
        cuda_device_count=cuda_device_count,
        cuda_backend=cuda_backend,
        xpu_available=xpu_available,
        xpu_device_count=xpu_device_count,
        mps_available=mps_available,
    )

    module_path = Path(__file__).resolve().parents[2] / "src" / "client.py"
    spec = importlib.util.spec_from_file_location("client_under_test", module_path)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    spec.loader.exec_module(module)
    return module, fake_torch

# ============================================================================
# BASIC DEVICE SELECTION TESTS
# ============================================================================

def test_select_device_force_cpu(monkeypatch):
    """Force CPU override takes precedence over all accelerators"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 7

    selected = client._select_device()
    assert selected == "cpu"


def test_select_device_prefers_npu_when_available(monkeypatch):
    """NPU is preferred device when available and enabled"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=True, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "1,2")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 3

    selected = client._select_device()
    assert selected == "npu:1"
    assert fake_torch.npu.selected_device == "npu:1"


def test_select_device_falls_back_to_cuda_on_npu_error(monkeypatch):
    """Falls back to CUDA/GPU when NPU throws error"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, npu_raises=True, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 9

    selected = client._select_device()
    assert selected == "cuda:0"


def test_select_device_falls_back_to_cpu_when_no_accelerator(monkeypatch):
    """Falls back to CPU when no accelerators available"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 11

    selected = client._select_device()
    assert selected == "cpu"


# ============================================================================
# MULTI-GPU DEVICE AFFINITY TESTS
# ============================================================================

def test_multi_npu_device_selection_preferred_device(monkeypatch):
    """Select first available device from ASCEND_RT_VISIBLE_DEVICES"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=4, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    # Test various ASCEND_RT_VISIBLE_DEVICES settings
    test_cases = [
        ("0", "npu:0"),
        ("1", "npu:1"),
        ("2", "npu:2"),
        ("3", "npu:3"),
        ("0,1,2,3", "npu:0"),  # First device in list
        ("3,2,1,0", "npu:3"),  # First device in list
    ]

    for visible_devices, expected_device in test_cases:
        monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", visible_devices)
        client = module.SovereignClient.__new__(module.SovereignClient)
        client.node_id = 0
        selected = client._select_device()
        assert selected == expected_device, f"With ASCEND_RT_VISIBLE_DEVICES={visible_devices}, expected {expected_device}, got {selected}"


def test_multi_cuda_device_selection(monkeypatch):
    """Select different CUDA devices when NPU unavailable"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True, cuda_device_count=2
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected in ["cuda:0", "cuda:1"]  # Should select a valid CUDA device


def test_npu_visible_devices_environment_variable(monkeypatch):
    """Respect ASCEND_RT_VISIBLE_DEVICES for device pool selection"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=8, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "2,5,7")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    # Should select from visible devices only
    assert selected in ["npu:2", "npu:5", "npu:7"]


# ============================================================================
# NPU DISABLED MODE TESTS
# ============================================================================

def test_npu_disabled_falls_back_to_cuda(monkeypatch):
    """When NPU_ENABLED=false, skip NPU and try CUDA"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 5

    selected = client._select_device()
    assert selected == "cuda:0", "Should use CUDA when NPU is disabled"


def test_npu_disabled_falls_back_to_cpu(monkeypatch):
    """When NPU_ENABLED=false and no CUDA, use CPU"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 10

    selected = client._select_device()
    assert selected == "cpu"


# ============================================================================
# XPU / ROCM / MPS SUPPORT TESTS
# ============================================================================

def test_xpu_selected_when_npu_unavailable(monkeypatch):
    """Select XPU when NPU is unavailable and XPU is present"""
    module, _ = _load_client_module(
        monkeypatch,
        npu_available=False,
        cuda_available=True,
        xpu_available=True,
        xpu_device_count=2,
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("XPU_ENABLED", "true")
    monkeypatch.setenv("XPU_VISIBLE_DEVICES", "1,0")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 4

    selected = client._select_device()
    assert selected == "xpu:1"


def test_xpu_disabled_falls_back_to_cuda(monkeypatch):
    """Skip XPU when disabled and use CUDA instead"""
    module, _ = _load_client_module(
        monkeypatch,
        npu_available=False,
        cuda_available=True,
        xpu_available=True,
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")
    monkeypatch.setenv("XPU_ENABLED", "false")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 2

    selected = client._select_device()
    assert selected == "cuda:0"


def test_xpu_visible_devices_environment_variable(monkeypatch):
    """Respect XPU_VISIBLE_DEVICES for XPU selection"""
    module, fake_torch = _load_client_module(
        monkeypatch,
        npu_available=False,
        cuda_available=False,
        xpu_available=True,
        xpu_device_count=4,
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")
    monkeypatch.setenv("XPU_ENABLED", "true")
    monkeypatch.setenv("XPU_VISIBLE_DEVICES", "2,3")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "xpu:2"
    assert fake_torch.xpu.selected_device == "xpu:2"


def test_rocm_backend_uses_cuda_device_namespace(monkeypatch):
    """AMD ROCm GPUs are selected through the CUDA namespace in PyTorch"""
    module, fake_torch = _load_client_module(
        monkeypatch,
        npu_available=False,
        cuda_available=True,
        cuda_backend="rocm",
        cuda_device_count=2,
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")
    monkeypatch.setenv("HIP_VISIBLE_DEVICES", "1,0")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "cuda:1"
    assert fake_torch.version.hip == "6.1"
    assert fake_torch.cuda.get_device_name(0).startswith("AMD Radeon Instinct MI")


def test_mps_selected_when_other_accelerators_unavailable(monkeypatch):
    """Use MPS when NPU, XPU, and CUDA are unavailable"""
    module, _ = _load_client_module(
        monkeypatch,
        npu_available=False,
        cuda_available=False,
        xpu_available=False,
        mps_available=True,
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")
    monkeypatch.setenv("XPU_ENABLED", "false")
    monkeypatch.setenv("GPU_ENABLED", "false")
    monkeypatch.setenv("MPS_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "mps"


# ============================================================================
# ENVIRONMENT VARIABLE EDGE CASES
# ============================================================================

def test_force_cpu_empty_string(monkeypatch):
    """FORCE_CPU empty string should be treated as false"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 1

    selected = client._select_device()
    # Empty string should not force CPU, should use accelerators
    assert selected != "cpu"


def test_visible_devices_with_single_device(monkeypatch):
    """ASCEND_RT_VISIBLE_DEVICES with single NPU"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=4, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "2")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "npu:2"


def test_visible_devices_with_all_devices(monkeypatch):
    """ASCEND_RT_VISIBLE_DEVICES with all available devices"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=4, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "0,1,2,3")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 10

    selected = client._select_device()
    assert selected in ["npu:0", "npu:1", "npu:2", "npu:3"]


# ============================================================================
# NPU EXCEPTION HANDLING TESTS
# ============================================================================

def test_npu_runtime_exception_fallback(monkeypatch):
    """NPU runtime exception should fallback to next device"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, npu_raises=True, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "cuda:0", "Should fallback to CUDA when NPU raises"


def test_npu_exception_cuda_unavailable(monkeypatch):
    """NPU exception with no CUDA should fallback to CPU"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, npu_raises=True, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "cpu"


# ============================================================================
# NODE ID DISTRIBUTION TESTS
# ============================================================================

def test_large_node_id_has_no_effect_on_device_selection(monkeypatch):
    """Large node IDs don't affect which NPU device is selected"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=2, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "0,1")

    # Device selection is determined by ASCEND_RT_VISIBLE_DEVICES, not node_id
    test_cases = [0, 1, 2, 100, 1000]
    
    for node_id in test_cases:
        client = module.SovereignClient.__new__(module.SovereignClient)
        client.node_id = node_id
        selected = client._select_device()
        assert selected == "npu:0", f"Node {node_id} should always use first visible device (npu:0)"


def test_sequential_node_ids_use_same_device(monkeypatch):
    """Sequential node IDs all use the same first visible device"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=4, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "0,1,2,3")

    # All nodes should select the same device (first in ASCEND_RT_VISIBLE_DEVICES)
    selected_devices = []
    for node_id in range(12):
        client = module.SovereignClient.__new__(module.SovereignClient)
        client.node_id = node_id
        selected = client._select_device()
        selected_devices.append(selected)

    # All should be the first visible device
    assert all(d == "npu:0" for d in selected_devices), f"All nodes should use npu:0, got {set(selected_devices)}"


# ============================================================================
# DEVICE PROPERTIES QUERY TESTS
# ============================================================================

def test_query_npu_device_properties(monkeypatch):
    """Query NPU device properties (memory, compute capability)"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=2, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    # Query device 0
    props = fake_torch.npu.get_device_properties(0)
    assert props.name == "Huawei Ascend NPU 0"
    assert props.total_memory == 80 * 1024 * 1024 * 1024  # 80GB


def test_query_cuda_device_properties(monkeypatch):
    """Query CUDA device properties (memory, compute capability)"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True, cuda_device_count=2
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")

    # Query device 0
    props = fake_torch.cuda.get_device_properties(0)
    assert props.name == "NVIDIA Tesla V100"
    assert props.total_memory == 24 * 1024 * 1024 * 1024  # 24GB


def test_query_cuda_device_name(monkeypatch):
    """Get CUDA device name"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True, cuda_device_count=2
    )

    name = fake_torch.cuda.get_device_name(0)
    assert name == "NVIDIA Tesla V100"

    name = fake_torch.cuda.get_device_name(1)
    assert name == "NVIDIA Tesla V101"


def test_device_count_npu(monkeypatch):
    """Query NPU device count"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=4, cuda_available=False
    )

    count = fake_torch.npu.device_count()
    assert count == 4


def test_device_count_cuda(monkeypatch):
    """Query CUDA device count"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True, cuda_device_count=2
    )

    count = fake_torch.cuda.device_count()
    assert count == 2


def test_device_count_unavailable(monkeypatch):
    """Device count is 0 when device type unavailable"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=False
    )

    npu_count = fake_torch.npu.device_count()
    cuda_count = fake_torch.cuda.device_count()

    assert npu_count == 0
    assert cuda_count == 0


# ============================================================================
# MEMORY CACHE CLEARING TESTS
# ============================================================================

def test_npu_empty_cache(monkeypatch):
    """NPU empty_cache() should be callable without error"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=1, cuda_available=False
    )

    # Should not raise
    fake_torch.npu.empty_cache()


def test_cuda_empty_cache(monkeypatch):
    """CUDA empty_cache() should be callable without error"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True, cuda_device_count=1
    )

    # Should not raise
    fake_torch.cuda.empty_cache()


# ============================================================================
# DEVICE FORMAT STRING TESTS
# ============================================================================

def test_device_string_format_npu(monkeypatch):
    """Device string format 'npu:N' is correctly formatted"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=4, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "0,1,2,3")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected.startswith("npu:")
    assert len(selected) > 4  # "npu:X"


def test_device_string_format_cuda(monkeypatch):
    """Device string format 'cuda:N' is correctly formatted"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True, cuda_device_count=2
    )
    monkeypatch.setenv("FORCE_CPU", "false")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected.startswith("cuda:")
    assert len(selected) > 5  # "cuda:X"


def test_device_string_format_cpu(monkeypatch):
    """Device string format 'cpu' is correctly formatted"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "cpu"


# ============================================================================
# ADDITIONAL COMPREHENSIVE DEVICE SELECTION TESTS
# ============================================================================

def test_npu_device_selection_with_probe_failure(monkeypatch):
    """Fall back to CUDA if NPU device probe fails"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    # Monkey patch _probe_device to simulate failure
    def mock_probe_npu(device):
        if "npu" in str(device):
            return False
        return True

    client._probe_device = mock_probe_npu
    selected = client._select_device()
    assert selected != "npu:0", "Should not use NPU if probe fails"


def test_npu_set_device_called_on_selection(monkeypatch):
    """When NPU is selected, torch.npu.set_device() is called"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=True, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    # Verify set_device was called
    assert fake_torch.npu.selected_device is not None
    assert "npu:" in selected


def test_cuda_selection_when_npu_unavailable_despite_enabled(monkeypatch):
    """Even with NPU_ENABLED=true, if NPU unavailable, use CUDA"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "cuda:0"


def test_empty_visible_devices_uses_default(monkeypatch):
    """Empty ASCEND_RT_VISIBLE_DEVICES should use default (device 0)"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=4, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    # Should default to device 0
    assert "npu:" in selected


def test_multiple_visible_npu_devices_selects_first(monkeypatch):
    """With multiple visible NPU devices, select the first one"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=8, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "5,6,7")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 100  # Large node_id shouldn't affect selection

    selected = client._select_device()
    assert selected == "npu:5"


def test_torch_version_property_available(monkeypatch):
    """torch.__version__ property is accessible"""
    module, fake_torch = _load_client_module(
        monkeypatch, npu_available=True, cuda_available=False
    )

    version = fake_torch.__version__
    assert version == "2.1.0"


def test_force_cpu_overrides_cuda(monkeypatch):
    """FORCE_CPU=true overrides CUDA even when available"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "cpu"


def test_force_cpu_case_insensitive(monkeypatch):
    """FORCE_CPU environment variable is case-insensitive"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "TRUE")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "cpu", "Should handle FORCE_CPU=TRUE (uppercase)"


def test_npu_enabled_case_insensitive(monkeypatch):
    """NPU_ENABLED environment variable is case-insensitive"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=False, cuda_available=True
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "FALSE")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    assert selected == "cuda:0", "When NPU_ENABLED=false, should use CUDA"


def test_visible_devices_with_spaces(monkeypatch):
    """ASCEND_RT_VISIBLE_DEVICES handles spaces around device IDs"""
    module, _ = _load_client_module(
        monkeypatch, npu_available=True, npu_device_count=4, cuda_available=False
    )
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", " 2 , 3 , 1 ")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 0

    selected = client._select_device()
    # Should strip spaces and use first device
    assert "npu:" in selected

