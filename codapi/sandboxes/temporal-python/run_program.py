# ABOUTME: Runs a learner's Python program, then exits without interpreter finalization.
# Works around temporalio/sdk-python#300, where dropping the SDK's Rust runtime at exit can hang under CPU contention.
"""Launch a Python file or module as ``__main__`` and exit with ``os._exit``."""

import os
import runpy
import sys
import traceback


def run(target: str, *, as_module: bool = False) -> int:
    """Run a Python file or module as ``__main__`` and report how it ended.

    Tracebacks start at the learner's file, hiding this launcher's frames.

    Args:
        target: The Python file to run, or the module name when ``as_module`` is set.
        as_module: Run ``target`` like ``python -m target`` instead of as a file path.

    Returns:
        The process exit code the program would have produced if run directly.
    """
    path = target
    # Match `python file.py` (script directory first on sys.path) and `python -m` (current directory first).
    sys.path.insert(0, os.getcwd() if as_module else os.path.dirname(os.path.abspath(target)))
    try:
        if as_module:
            runpy.run_module(target, run_name="__main__", alter_sys=True)
        else:
            runpy.run_path(target, run_name="__main__")
    except SystemExit as exit_request:
        match exit_request.code:
            case None:
                return 0
            case int() as code:
                return code
            case message:
                print(message, file=sys.stderr)
                return 1
    except BaseException as error:
        # Skip frames until the learner's file so the traceback matches `python main.py`.
        frame = error.__traceback__
        while frame is not None and frame.tb_frame.f_code.co_filename != path:
            frame = frame.tb_next
        traceback.print_exception(type(error), error, frame)
        return 1
    return 0


def main() -> None:
    """Run ``<file> [args...]`` or ``-m <module> [args...]`` and exit with its status."""
    as_module = sys.argv[1] == "-m"
    sys.argv = sys.argv[2:] if as_module else sys.argv[1:]
    exit_code = run(sys.argv[0], as_module=as_module)

    # os._exit skips the interpreter's normal shutdown, including flushing output buffers.
    sys.stdout.flush()
    sys.stderr.flush()
    os._exit(exit_code)


if __name__ == "__main__":
    main()
