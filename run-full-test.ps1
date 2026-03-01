$ScriptPath = Join-Path $PSScriptRoot "tests/scripts/powershell/run-full-test.ps1"
& $ScriptPath @args
exit $LASTEXITCODE
