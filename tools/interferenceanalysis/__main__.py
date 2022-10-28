import argparse

from . import Preprocess, StressProcess, MachineProcess


def preprocess_main(args):
    preprocess = Preprocess(args.metrics, args.qos, args.output)
    preprocess.execute()


def stress_process_main(args):
    stress_process = StressProcess(args.stress, args.qos, args.output)
    stress_process.execute()

def machine_process_main(args):
    machine_process = MachineProcess(args.stress, args.machine, args.output)
    machine_process.execute()

def main() -> None:
    """
    The main function executes on commands:
    `python -m rubikanalysis` and `$ rubikanalysis `.

    This is your program's entry point.
    """
    parser = argparse.ArgumentParser(
        description="A tool for metrics analysis of co-Location container workloads",
        epilog=("Through the software-hardware collaborative analysis method, the application"
                " characteristics and system-level characteristics of load execution under the"
                " conditions of differnet configurations of clusters and different co-location"
                " modes are studied to guide the resource planning and scheduling management of"
                " cloud clusters"),
    )

    subparsers = parser.add_subparsers()
    # create the parser for the "preprocess" command
    parser_preprocess = subparsers.add_parser(
        'preprocess', help='Indicator data preprocessing')
    parser_preprocess.add_argument(
        '-m', "--metrics", type=str, help='Low level metrics', required=True)
    parser_preprocess.add_argument(
        '-q', "--qos", type=str, help='High level SLAs', required=True)
    parser_preprocess.add_argument(
        '-o', "--output", type=str, help='Preprocessed merged file', required=True)
    parser_preprocess.set_defaults(func=preprocess_main)

    parser_stress_process = subparsers.add_parser(
        'stress-process', help='Stress data processing')
    parser_stress_process.add_argument(
        '-s', "--stress", type=str, help='Stress info file', required=True)
    parser_stress_process.add_argument(
        '-q', "--qos", type=str, help='High level SLAs', required=True)
    parser_stress_process.add_argument(
        '-o', "--output", type=str, help='Stress merged file', required=True)
    parser_stress_process.set_defaults(func=stress_process_main)

    parser_machine_process = subparsers.add_parser(
        'machine-process', help='Machine data processing')
    parser_machine_process.add_argument(
        '-s', "--stress", type=str, help='Stress info file', required=True)
    parser_machine_process.add_argument(
        '-m', "--machine", type=str, help='Machine data file', required=True)
    parser_machine_process.add_argument(
        '-o', "--output", type=str, help='Merged file', required=True)
    parser_machine_process.set_defaults(func=machine_process_main)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
