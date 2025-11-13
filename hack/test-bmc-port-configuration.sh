#!/usr/bin/env bash

set -e
set -o pipefail

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"

echo "Running BMC port configuration tests..."
echo "========================================="

echo ""
echo "1. Running unit tests for BMC port functionality..."
go test -v "${SCRIPT_ROOT}/pkg/providers/tinkerbell/hardware/..." \
    -run "TestBMCCatalogueWriter_WriteWith|TestStaticMachineAssertions.*BMCPort"

echo ""
echo "2. Testing CSV parsing with bmc_port column..."
go test -v "${SCRIPT_ROOT}/pkg/providers/tinkerbell/hardware/..." \
    -run "TestCSVReader"

echo ""
echo "3. Running all hardware validation tests..."
go test -v "${SCRIPT_ROOT}/pkg/providers/tinkerbell/hardware/..." \
    -run "TestDefaultMachineValidator|TestStaticMachineAssertions"

echo ""
echo "========================================="
echo "All BMC port configuration tests passed!"

