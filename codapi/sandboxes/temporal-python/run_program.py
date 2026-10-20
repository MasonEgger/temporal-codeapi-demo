# ABOUTME: Runs a learner's Python program, then exits without interpreter finalization.
# Works around temporalio/sdk-python#300, where dropping the SDK's Rust runtime at exit can hang under CPU contention.
"""Launch a Python file as ``__main__`` and exit with ``os._exit``."""

import os
import runpy
import sys
import traceback


def run(path: str) -> int:
    """Run a Python file as ``__main__`` and report how it ended.

    Tracebacks start at the learner's file, hiding this launcher's frames.

    Args:
        path: The Python file to run.

    Returns:
        The process exit code the program would have produced if run directly.
    """
    try:
        runpy.run_path(path, run_name="__main__")
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
    """Run the file named by the first argument and exit with its status."""
    sys.argv = sys.argv[1:]
    exit_code = run(sys.argv[0])

    # os._exit skips the interpreter's normal shutdown, including flushing output buffers.
    sys.stdout.flush()
    sys.stderr.flush()
    os._exit(exit_code)


if __name__ == "__main__":
    main()
