import yaml

def read_template_files(template_first, template_second):
    """
    Read the template and params files from the user-specified paths.
    """
    with open(template_first, 'r') as template_first_file, \
         open(template_second, 'r') as template_second_file:
        t1 = yaml.load(template_first_file, Loader=yaml.FullLoader)
        t2 = yaml.load(template_second_file, Loader=yaml.FullLoader)
        return t1, t2


def compare_keys(yaml1, yaml2, path=""):
    keys1, keys2 = set(yaml1.keys()), set(yaml2.keys())

    #print(f"Keys in template1: {keys1}")
    #print(f"Keys in template2: {keys2}")
    
    
    added_keys = keys2 - keys1
    deleted_keys = keys1 - keys2
    altered_keys = {k for k in (keys1 & keys2) if isinstance(yaml1[k], dict) and isinstance(yaml2[k], dict)}
    
    for key in altered_keys:
        compare_keys(yaml1[key], yaml2[key], path + key + ".")
    
    if added_keys:
        print(f"Added keys in {path}: {added_keys}")
    if deleted_keys:
        print(f"Deleted keys in {path}: {deleted_keys}")

if __name__ == "__main__":
    import sys
    template_first = input("Please enter the first template file: ")
    template_second = input("Please enter the second template file: ")

    t1, t2 = read_template_files(template_first, template_second)

    compare_keys(t1, t2)


