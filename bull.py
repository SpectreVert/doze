import argparse
import logging
from pathlib import Path
import sys

def verify_layout(layout, basedir):
    if not basedir.exists():
        logging.error(f"{basedir} not found")
        raise SystemExit(1)
    verified = 0
    checked = 0
    total_length = 0
    for role, items in layout.items():
        total_length += len(items)
        for item_to_check in items:
            if basedir / Path(item_to_check) in basedir.iterdir():
                logging.info(f"({role}) {item_to_check}")
                verified += 1
            else:
                logging.error(f"({role}) {item_to_check} NOT FOUND")
            checked += 1
    if checked != verified:
        logging.error(f"verify_layout: only verifed {verifed} items out of {checked}")
        return False
    present_files = 0
    for file in basedir.iterdir():
        present_files += 1
    if present_files != total_length:
        logging.error(f"verify_layout: incorrect amount of files in {basedir}")
        return False
    return True


def load_layout(layout_file):
    if not layout_file.exists():
        logging.error(f"{layout_file} not found")
        raise SystemExit(1)
    layout = {
        "i": [],
        "a": [],
    }
    data = layout_file.read_text()
    for line in data.splitlines():
        role, tag = line.split(" ")
        layout[role].append(tag)
    return layout


def setup_parser():
    parser = argparse.ArgumentParser()

    parser.add_argument("--in", required=True, dest="input", type=Path, help="initial layout map")
    parser.add_argument("--out", required=True, dest="output", type=Path, help="final layout map")
    parser.add_argument("--dir-in", required=True, dest="dir_in", type=Path, help="directory with the real files")
    parser.add_argument("--dir-out", required=True, dest="dir_out", type=Path, help="directory with the real files")

    return parser


def main():
    logging.basicConfig(
        stream=sys.stdout,
        format="%(asctime)s %(levelname)s %(message)s",
        level=logging.INFO,
    )

    parser = setup_parser()
    args = parser.parse_args()

    try:
        inputs = load_layout(args.input)
        if not verify_layout(inputs, args.dir_in):
            raise SystemExit(1)
        print(inputs)

        print("run doze")

        outputs = load_layout(args.output)
        if not verify_layout(outputs, args.dir_out):
            raise SystemExit(1)
        print(outputs)

        # verify_layout(args.out)
    except SystemExit as e:
        return 1


if __name__ == "__main__":
    sys.exit(main())
