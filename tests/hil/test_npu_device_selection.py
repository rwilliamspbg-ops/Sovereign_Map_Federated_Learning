import importlib.util
import sys
import types
from pathlib import Path


class _FakeNPU:
    def __init__(self, available: bool = False, raise_on_check: bool = False):
        self._available = available
        self._raise_on_check = raise_on_check
        self.selected_device = None

    def is_available(self) -> bool:
        if self._raise_on_check:
            raise RuntimeError("NPU runtime unavailable")
        return self._available

    def set_device(self, device):
        self.selected_device = device


class _FakeCUDA:
    def __init__(self, available: bool):
        self._available = available

    def is_available(self) -> bool:
        return self._available


def _install_module_stubs(monkeypatch, *, npu_available=False, npu_raises=False, cuda_available=False):
    fake_torch = types.ModuleType("torch")
    fake_torch.npu = _FakeNPU(available=npu_available, raise_on_check=npu_raises)
    fake_torch.cuda = _FakeCUDA(available=cuda_available)
    fake_torch.device = lambda name: name

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
    fake_torchvision.transforms = types.SimpleNamespace(Compose=lambda *args, **kwargs: None)

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


def _load_client_module(monkeypatch, *, npu_available=False, npu_raises=False, cuda_available=False):
    fake_torch = _install_module_stubs(
        monkeypatch,
        npu_available=npu_available,
        npu_raises=npu_raises,
        cuda_available=cuda_available,
    )

    module_path = Path(__file__).resolve().parents[2] / "src" / "client.py"
    spec = importlib.util.spec_from_file_location("client_under_test", module_path)
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    spec.loader.exec_module(module)
    return module, fake_torch


def test_select_device_force_cpu(monkeypatch):
    module, _ = _load_client_module(monkeypatch, npu_available=True, cuda_available=True)
    monkeypatch.setenv("FORCE_CPU", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 7

    selected = client._select_device()
    assert selected == "cpu"


def test_select_device_prefers_npu_when_available(monkeypatch):
    module, fake_torch = _load_client_module(monkeypatch, npu_available=True, cuda_available=True)
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")
    monkeypatch.setenv("ASCEND_RT_VISIBLE_DEVICES", "1,2")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 3

    selected = client._select_device()
    assert selected == "npu:1"
    assert fake_torch.npu.selected_device == "npu:1"


def test_select_device_falls_back_to_cuda_on_npu_error(monkeypatch):
    module, _ = _load_client_module(monkeypatch, npu_available=False, npu_raises=True, cuda_available=True)
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "true")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 9

    selected = client._select_device()
    assert selected == "cuda:0"


def test_select_device_falls_back_to_cpu_when_no_accelerator(monkeypatch):
    module, _ = _load_client_module(monkeypatch, npu_available=False, cuda_available=False)
    monkeypatch.setenv("FORCE_CPU", "false")
    monkeypatch.setenv("NPU_ENABLED", "false")

    client = module.SovereignClient.__new__(module.SovereignClient)
    client.node_id = 11

    selected = client._select_device()
    assert selected == "cpu"