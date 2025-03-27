import yaml
import difflib

def compare_yaml(file1, file2):
    with open(file1, 'r') as f1, open(file2, 'r') as f2:
        yaml1 = yaml.load(f1)
        yaml2 = yaml.load(f2)
    d = difflib.Difference()
    differ = difflib.Differ()
    diff = list(differ.compare(yaml.dump(yaml1, default_flow_style=False).splitlines(True),
                              yaml.dump(yaml2, default_flow_style=False).splitlines(True)))
    for line in diff:
        if line.startswith('+'):
            print(line)
        elif line.startswith('-'):
            print(line)

if __name__ == "__main__":
    import sys
    compare_yaml(sys.argv[1], sys.argv[2])
