# secrets/

sops-encrypted files, encrypted to the age recipients in `../.sops.yaml`.
Plaintext exists only in `$EDITOR` while editing and in the environment of the one `terraform` process that `just infra-*` runs.
The ciphertext is safe in a public repository.

| File | Holds | Consumed by |
|---|---|---|
| `terraform.sops.env` | `TF_VAR_do_token` (DigitalOcean API token) | `sops exec-env` around every Terraform run (`_tf` in the justfile) |

Edit with `just sops secrets/terraform.sops.env` from the repo root.
Read without editing: `sops -d secrets/terraform.sops.env`.
After adding or removing a recipient in `.sops.yaml`, run `just rekey`.
Nothing here is ever copied to the droplet; Ansible needs no secrets.
