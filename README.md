### building

This allows for cross compiling to arm64 when using MacOS as a dev platform:

```bash
docker build -t pam-builder .
docker create --name extract pam-builder
docker cp extract:/src/supa_jitdb_pam.so ./pam_jit_pg.so
docker rm extract
```

Build with nix (using docker as a standin here for a linux host):

```bash
docker build -t nix-go-pam -f Dockerfile_nix .
docker create --name temp nix-go-pam
docker cp temp:/app/result/lib/security/pam_jit_pg.so ./pam_jwt_pg_nix.so
docker rm temp
```

### setup on the server

Copy the `.so` to the server. And add to the correct pam location, normally:

```
/lib/aarch64-linux-gnu/security/
```

In the case of `nix` builds, such as the Supabase image, it needs to go the nix store:

```
cp pam_jit_pg.so /nix/store/*-linux-pam-1.6.0/lib/security/
```

Next setup `/etc/pam.d/postgresql` with the following

```
auth required pam_jit_pg.so jwks=https://auth.supabase.green/auth/v1/.well-known/jwks.json mappings=/tmp/users.yaml
account required pam_jit_pg.so jwks=https://auth.supabase.green/auth/v1/.well-known/jwks.json mappings=/tmp/users.yaml
```

The `apiUrl` value should point to the URL of a valid api that accepts the PAT and/or JWT for authentication. The API should return a JSON struct with the roles the user associated to the PAT/JWT is allowed to assume:

```

{
  "user_id":"2256c8fe-95a6-4554-a2e3-0e6a095b72d7",
  "user_roles":
    [
      { "role":"postgres",
        "expires_at":"1753280418791"
      }
    ]
}
```


Finally setup the pg_hba.conf:

```
host  all  postgres  ::0/0     scram-sha-256
host all  all 0.0.0.0/0 pam
host all  all ::0/0 pam
```

Reload postgresql.service:

```
systemctl reload postgresql
```

And now login with JWT should work, as long as the JWT is signed by a key found in the jwks URL, the user email in the JWT matches one in the mappings file, the chosen postgres user role is permitted to the user and the JWT is still valid.
