import os
import sys
import platform
import subprocess
from pathlib import Path

def get_binary_path():
    """Get the appropriate binary for the current platform and architecture"""
    system = platform.system().lower()
    machine = platform.machine().lower()

    # Normalize architecture names
    if machine in ['x86_64', 'amd64']:
        arch = 'x86_64'
    elif machine in ['aarch64', 'arm64']:
        arch = 'arm64'
    else:
        raise RuntimeError(f"Unsupported architecture: {machine}")

    # Determine platform and binary name
    if system == 'linux':
        platform_name = 'linux'
        binary_name = 'ckc'
    elif system == 'darwin':
        platform_name = 'darwin'
        binary_name = 'ckc'
    elif system == 'windows':
        platform_name = 'windows'
        binary_name = 'ckc.exe'
    else:
        raise RuntimeError(f"Unsupported operating system: {system}")

    # Build path to binary
    base_dir = Path(__file__).parent
    binary_path = base_dir / "bin" / arch / platform_name / binary_name

    if not binary_path.exists():
        # Fallback: try to find any available binary
        available_binaries = []
        bin_dir = base_dir / "bin"
        if bin_dir.exists():
            for arch_dir in bin_dir.iterdir():
                if arch_dir.is_dir():
                    for platform_dir in arch_dir.iterdir():
                        if platform_dir.is_dir():
                            for binary in platform_dir.iterdir():
                                if binary.is_file():
                                    available_binaries.append(str(binary.relative_to(base_dir)))

        error_msg = f"Binary not found for {system}/{machine} at: {binary_path.relative_to(base_dir)}"
        if available_binaries:
            error_msg += f"\nAvailable binaries: {', '.join(available_binaries)}"

        raise RuntimeError(error_msg)

    return binary_path

def main():
    """Main entry point that calls the appropriate Go binary"""
    try:
        binary_path = get_binary_path()

        # Make sure binary is executable on Unix systems
        if os.name != 'nt':
            os.chmod(binary_path, 0o755)

        # Execute the binary with all passed arguments
        result = subprocess.run([str(binary_path)] + sys.argv[1:])
        sys.exit(result.returncode)

    except RuntimeError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except KeyboardInterrupt:
        sys.exit(130)  # Standard exit code for Ctrl+C
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
