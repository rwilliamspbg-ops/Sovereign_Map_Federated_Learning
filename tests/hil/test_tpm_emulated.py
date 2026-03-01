import os
import shutil
import socket
import subprocess
import time
from pathlib import Path

import pytest


def _find_free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind(("127.0.0.1", 0))
        return int(sock.getsockname()[1])


def _find_free_consecutive_ports() -> tuple[int, int]:
    for _ in range(50):
        base = _find_free_port()
        ctrl = base + 1
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as first, socket.socket(
            socket.AF_INET, socket.SOCK_STREAM
        ) as second:
            try:
                first.bind(("127.0.0.1", base))
                second.bind(("127.0.0.1", ctrl))
                return base, ctrl
            except OSError:
                continue
    raise RuntimeError("Unable to allocate consecutive free ports for swtpm")


def _run_cmd(
    command: list[str], env: dict[str, str] | None = None, cwd: Path | None = None
) -> subprocess.CompletedProcess:
    return subprocess.run(
        command,
        check=True,
        capture_output=True,
        text=True,
        env=env,
        cwd=str(cwd) if cwd else None,
    )


@pytest.fixture(scope="module")
def swtpm_env(tmp_path_factory: pytest.TempPathFactory) -> dict[str, str]:
    required_bins = [
        "swtpm",
        "swtpm_setup",
        "tpm2_getcap",
        "tpm2_createprimary",
        "tpm2_pcrread",
        "tpm2_getrandom",
    ]
    missing = [binary for binary in required_bins if shutil.which(binary) is None]
    if missing:
        pytest.skip(f"Missing required TPM tools: {', '.join(missing)}")

    work_dir = tmp_path_factory.mktemp("swtpm-state")
    tpm_state_dir = work_dir / "state"
    tpm_state_dir.mkdir(parents=True, exist_ok=True)

    command_port, control_port = _find_free_consecutive_ports()

    _run_cmd(
        [
            "swtpm_setup",
            "--tpmstate",
            str(tpm_state_dir),
            "--tpm2",
        ]
    )

    process = subprocess.Popen(
        [
            "swtpm",
            "socket",
            "--tpmstate",
            f"dir={tpm_state_dir}",
            "--tpm2",
            "--server",
            f"type=tcp,port={command_port}",
            "--ctrl",
            f"type=tcp,port={control_port}",
            "--flags",
            "not-need-init,startup-clear",
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )

    env = os.environ.copy()
    env["TPM2TOOLS_TCTI"] = f"swtpm:host=127.0.0.1,port={command_port}"

    last_error: Exception | None = None
    for _ in range(20):
        try:
            _run_cmd(["tpm2_getcap", "properties-fixed"], env=env)
            last_error = None
            break
        except Exception as exc:
            last_error = exc
            time.sleep(0.5)

    if last_error is not None:
        process.terminate()
        process.wait(timeout=5)
        raise RuntimeError("swtpm failed to start in time") from last_error

    env["HIL_SWTPM_WORKDIR"] = str(work_dir)

    try:
        yield env
    finally:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()


def test_tpm2_getcap_properties_fixed(swtpm_env: dict[str, str]) -> None:
    result = _run_cmd(["tpm2_getcap", "properties-fixed"], env=swtpm_env)
    assert "TPM2_PT_FAMILY_INDICATOR" in result.stdout


def test_tpm2_can_create_primary_and_read_pcr(swtpm_env: dict[str, str]) -> None:
    work_dir = Path(swtpm_env["HIL_SWTPM_WORKDIR"]) / "primary"
    work_dir.mkdir(parents=True, exist_ok=True)

    _run_cmd(
        ["tpm2_createprimary", "-C", "o", "-G", "rsa", "-c", "primary.ctx"],
        env=swtpm_env,
        cwd=work_dir,
    )
    result = _run_cmd(["tpm2_pcrread", "sha256:0"], env=swtpm_env, cwd=work_dir)

    assert "sha256:" in result.stdout


def test_tpm2_can_get_random_bytes(swtpm_env: dict[str, str]) -> None:
    work_dir = Path(swtpm_env["HIL_SWTPM_WORKDIR"]) / "random"
    work_dir.mkdir(parents=True, exist_ok=True)

    output_file = work_dir / "random.bin"
    _run_cmd(
        ["tpm2_getrandom", "16", "-o", str(output_file)], env=swtpm_env, cwd=work_dir
    )

    assert output_file.exists()
    assert output_file.stat().st_size == 16
