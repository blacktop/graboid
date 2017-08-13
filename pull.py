import requests
import json

namespace = "blacktop"
repo_name = "bro"
manifest_json = [{
    "Config": "", 
    "RepoTags": [f"{namespace}/{repo_name}:latest"],
    "Layers": []
}]


def remove_prefix(text, prefix):
    if text.startswith(prefix):
        return text[len(prefix):]
    return text


# GETTING TOKEN
url = f"https://auth.docker.io/token?service=registry.docker.io&scope=repository:{namespace}/{repo_name}:pull"
response = requests.get(url)
token = response.json().get("token")

# GETTING TAG LIST
header = {"Authorization": "Bearer " + token}
url = f"https://index.docker.io/v2/{namespace}/{repo_name}/tags/list"
response = requests.get(url, headers=header)
print(json.dumps(response.json(), sort_keys=True, indent=4))

# PULLING AN IMAGE MANIFEST
header = {"Authorization": "Bearer " + token, "Accept": "application/vnd.docker.distribution.manifest.v2+json"}
url = f"https://index.docker.io/v2/{namespace}/{repo_name}/manifests/latest"
response = requests.get(url, headers=header)
resp = response.json()
print(json.dumps(resp, sort_keys=True, indent=4))

# PULLING CONFIG
config = resp.get("config")
digest = config.get("digest")
header = {"Authorization": "Bearer " + token, "Accept": config.get("mediaType")}
url = f"https://index.docker.io/v2/{namespace}/{repo_name}/blobs/{digest}"
response = requests.get(url, headers=header, stream=True)
response.raise_for_status()
config_json = remove_prefix(digest, "sha256:") + ".json"
manifest_json[0]["Config"] = config_json
with open(config_json, 'wb') as handle:
    for block in response.iter_content(1024):
        handle.write(block)

# PULLING A LAYERS
for layer in resp.get("layers"):
    digest = layer.get("digest")
    header = {"Authorization": "Bearer " + token, "Accept": layer.get("mediaType")}
    url = f"https://index.docker.io/v2/{namespace}/{repo_name}/blobs/{digest}"
    response = requests.get(url, headers=header, stream=True)
    response.raise_for_status()
    digest_tar = remove_prefix(digest, "sha256:") + ".tar"
    manifest_json[0]["Layers"].append(digest_tar)
    with open(digest_tar, 'wb') as handle:
        for block in response.iter_content(1024):
            handle.write(block)

# CREATE MANIFEST
with open("manifest.json", 'w') as m:
    json.dump(manifest_json, m)
