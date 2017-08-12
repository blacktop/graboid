import requests
import json

namespace = "library"
repo_name = "alpine"


def remove_prefix(text, prefix):
    if text.startswith(prefix):
        return text[len(prefix):]
    return text


#############################
# GETTING TOKEN             #
#############################
url = f"https://auth.docker.io/token?service=registry.docker.io&scope=repository:{namespace}/{repo_name}:pull"
response = requests.get(url)
token = response.json().get("token")

#############################
# PULLING AN IMAGE MANIFEST #
#############################
header = {"Authorization": "Bearer " + token, "Accept": "application/vnd.docker.distribution.manifest.v2+json"}
url = f"https://index.docker.io/v2/{namespace}/{repo_name}/manifests/latest"
response = requests.get(url, headers=header)
with open("manifest.json", 'wb') as handle:
    for block in response.iter_content(1024):
        handle.write(block)
resp = response.json()
print(json.dumps(resp, sort_keys=True, indent=4))

#############################
# PULLING CONFIG            #
#############################
config = resp.get("config")
digest = config.get("digest")
header = {"Authorization": "Bearer " + token, "Accept": config.get("mediaType")}
url = f"https://index.docker.io/v2/{namespace}/{repo_name}/blobs/{digest}"
response = requests.get(url, headers=header, stream=True)
response.raise_for_status()
with open(remove_prefix(digest, "sha256:") + ".json", 'wb') as handle:
    for block in response.iter_content(1024):
        handle.write(block)

#############################
# PULLING A LAYERS          #
#############################
for layer in resp.get("layers"):
    digest = layer.get("digest")
    header = {"Authorization": "Bearer " + token, "Accept": layer.get("mediaType")}
    url = f"https://index.docker.io/v2/{namespace}/{repo_name}/blobs/{digest}"
    response = requests.get(url, headers=header, stream=True)
    response.raise_for_status()
    with open(remove_prefix(digest, "sha256:") + ".tar.gzip", 'wb') as handle:
        for block in response.iter_content(1024):
            handle.write(block)

# for layer in response.json().get("fsLayers"):
#     digest = layer.get("blobSum")
#     header = {"Authorization": "Bearer " + token}
#     url = f"https://index.docker.io/v2/{namespace}/{repo_name}/blobs/{digest}"
#     response = requests.get(url, headers=header, stream=True)
#     response.raise_for_status()
#     with open(digest + ".tar.gz", 'wb') as handle:
#         for block in response.iter_content(1024):
#             handle.write(block)
