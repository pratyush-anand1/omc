import os
import yaml


def check_directory_structure(directory, required_files):
    for file in required_files:
        if not os.path.exists(os.path.join(directory, file)):
            print(f"Error: {file} is missing in {directory}")
            return False
    return True

def extract_yaml_files(directory, meDetails):
  #templateParameters:
  #   resourceParams:
  #     managed_element:
  #     single-server-configuration: 
  #   clusterParams:
  #     ccd_env:
  #     params: 
  #     user_secrets:
  
  #├── ccd_env.yaml
  #├── cluster_config
  #│   └── input
  #│       └── params.yaml
  #├── single-server-configuration.yaml
  #└── user-secrets.yaml

    SINGLE_SERVER_CONFIGURATION_YAML = "single-server-configuration.yaml"
    CCD_ENV_YAML = "ccd_env.yaml"
    CLUSTER_CONFIG_INPUT_PARAMS_YAML = "cluster_config/input/params.yaml"
    USER_SECRETS_YAML = "user-secrets.yaml"


    required_files = [CCD_ENV_YAML, CLUSTER_CONFIG_INPUT_PARAMS_YAML,
                   SINGLE_SERVER_CONFIGURATION_YAML, USER_SECRETS_YAML]

    if not check_directory_structure(directory, required_files):
        exit(1)

    yaml_files = {}
    for file in required_files:
      with open(os.path.join(directory, file), 'r') as yaml_file:
          try:
              yaml_files[file] = yaml.safe_load(yaml_file)
              # fix for CLUSTER_CONFIG_INPUT_PARAMS_YAML
          except yaml.YAMLError as err:
              print(f"Error: {file} is not a valid yaml file: {err}")
              exit(1)
    print(yaml_files)
    print("formatting")
    template_params = {}
    template_params["templateParameters"] = {}
    template_params["templateParameters"]["resourceParams"]  = {} 
    template_params["templateParameters"]["resourceParams"]["managed_element"] = meDetails
    template_params["templateParameters"]["resourceParams"]["single-server-configuration"] = yaml_files[SINGLE_SERVER_CONFIGURATION_YAML]
    template_params["templateParameters"]["clusterParams"]  = {}
    template_params["templateParameters"]["clusterParams"]["ccd_env"]  = yaml_files[CCD_ENV_YAML]

    if "params" not in yaml_files[CLUSTER_CONFIG_INPUT_PARAMS_YAML]:
        print(f"Error: {CLUSTER_CONFIG_INPUT_PARAMS_YAML} does not contain a 'params' field")
        exit(1)
    template_params["templateParameters"]["clusterParams"]["params"]  = yaml_files[CLUSTER_CONFIG_INPUT_PARAMS_YAML]["params"]
    template_params["templateParameters"]["clusterParams"]["user_secrets"] = yaml_files[USER_SECRETS_YAML]
    return template_params
    
    with open(template_param_file, 'w') as outfile:
        yaml.dump(template_params, outfile, default_flow_style=False, sort_keys=True)
  

def generate_template_param_file(template_params, template_param_file):
    with open(template_param_file, 'w') as outfile:
        yaml.dump(template_params, outfile, default_flow_style=False, sort_keys=True)

if __name__ == "__main__":
    import sys
    default_template_param_file = './templateParams.yaml'
    default_configset = "./configset"
    # default values for managed element
    default_me_values = {
        'product': 'CNIS',
        'type': 'single-server',
        'software_version': '1.15'
    }

    default_crd_info = {
        'apiVersion': "o2ims.provisioning.oran.org/v1alpha1",
        'kind': "ProvisioningRequest",
        'spec': {
            'templateName': "single-node-lpg2",
            'templateVersion': "cnis-1.15_v1"
        }
    }

    me_name = input(f'Enter the value for managed element name(default=None): ')
    if not me_name:
        print("Error: managed element name must be provided")
        exit(1)

    me_description = input(f'Enter the value for managed element description(default=None): ')
    if not me_description:
        print("Error: managed element description must be provided")
        exit(1)
    default_crd_info['metadata'] = {
        'name': me_name
    }
    default_crd_info['spec']['description'] = me_description
    
    config_set_path = input(f"Please provide the path to the config set(default={default_configset}): ") or default_configset
    template_param_file = input(f"Enter the value for template param file(default={default_template_param_file}): ") or default_template_param_file
    me_values = {}
    for key, value in default_me_values.items():
        me_values[key] = input(f'Enter the value for {key} (default={value}): ') or value
    template_params = extract_yaml_files(config_set_path, me_values)
    generate_template_param_file(template_params, template_param_file)

    default_crd_info['spec']['templateParameters'] = template_params['templateParameters']

    path, filename = os.path.split(template_param_file)
    filename = os.path.splitext(filename)[0]
    crd_file_path = os.path.join(path, filename + '_crd' + os.path.splitext(filename)[1]+'.yaml')
    generate_template_param_file(default_crd_info, crd_file_path)



    
    

    
