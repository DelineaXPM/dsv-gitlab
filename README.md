# Repo

[![codecov](https://codecov.io/gh/delineaxpm/dsv-gitlab/branch/main/graph/badge.svg?token=FPHYYO5ZF2)](https://codecov.io/gh/delineaxpm/dsv-gitlab)

Delinea DevOps Secrets Vault (DSV) CI plugin allows you to access and reference your Secrets data available for use in GitLab Jobs.

## Getting Started

- [Developer](DEVELOPER.md): instructions on running tests, local tooling, and other resources.
- [DSV Documentation](https://docs.delinea.com/dsv/current?ref=githubrepo)

## Using With Gitlab

Review the file: [.gitlab-ci.yml](examples/.gitlab-ci.yml)

To test this out, you'll have to create variables in GitLab under: `https://gitlab.com/{org}/{project}/-/settings/ci_cd`.

## Prerequisites

This plugin uses authentication based on Client Credentials, i.e. via Client ID and Client Secret.

```shell
dsvprofile=

rolename="gitlab-dsv-gitlab-tests"
secretpath="ci:tests:dsv-gitlab"
secretpathclient="clients:${secretpath}"

desc="a secret for testing operation of secrets against dsv-gitlab"
clientcredfile=".cache/${rolename}.json"
clientcredname="${rolename}"

dsv role create --name "${rolename}" --profile $dsvprofile

# Option 1: Less Optimal - Save Credential to local json for testing
# dsv client create --role "${rolename}" --out "file:${clientcredfile}"

# Option 2: 🔒 MOST SECURE
# Create credential info for dsv, and set as variable.
# Create an org secret instead if you want to share this credential in many repos.

# compress to a single line
clientcred=$(dsv client create --role "${rolename}" --plain | jq -c)

# configure the credentials in gitlab
echo 'DSV_SERVER in GitLab variables, example: mytenant.secretsvaultcloud.com'
echo "Save DSV_CLIENT_ID in GitLab variables: $(echo "${clientcred}" | jq '.clientId' -r)"
echo "Save DSV_CLIENT_SECRET in GitLab variables: $(echo "${clientcred}" | jq '.clientSecret' -r )"
```

For further setup, here's how you could extend that script block above with also creating a secret and the policy to read just this secret.

```shell
# Create a secret
secretkey="secret-01"
secretvalue='{"value1":"taco","value2":"burrito"}'
dsv secret create \
  --path "secrets:${secretpath}:${secretkey}" \
  --data "${secretvalue}" \
  --desc "${desc}"

# Create a policy to allow role "$rolename" to read secrets under "ci:tests:integration-configs/dsv-gitlab":
dsv policy create \
  --path "secrets:${secretpath}" \
  --actions 'read' \
  --effect 'allow' \
  --subjects "roles:$rolename" \
  --desc "${desc}" \
  --resources "secrets:${secretpath}:<.*>"
```

## Usage

See [integration.yml](examples/.gitlab-ci.yml) for an example of how to use this to retrieve secrets and use outputs on other tasks.

### Retrieve 2 Values from Same Secret

The json expects an array, so just add a new line.

```yaml
retrieve: |
  [
   {"secretPath": "ci:tests:dsv-github-action:secret-01", "secretKey": "value1", "outputVariable": "RETURN_VALUE_1"},
   {"secretPath": "ci:tests:dsv-github-action:secret-01", "secretKey": "value2", "outputVariable": "RETURN_VALUE_2"}
  ]
```

### Retrieve 2 Values from Different Secrets

> Note: Make sure your generated client credentials are associated a policy that has rights to read the different secrets.

```yaml
retrieve: |
  [
   {"secretPath": "ci:tests:dsv-github-action:secret-01", "secretKey": "value1", "outputVariable": "RETURN_VALUE_1"},
   {"secretPath": "ci:tests:dsv-github-action:secret-02", "secretKey": "value1", "outputVariable": "RETURN_VALUE_2"}
  ]
```

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center"><a href="https://github.com/mariiatuzovska"><img src="https://avatars.githubusercontent.com/u/41679258?v=4?s=100" width="100px;" alt="Mariia"/><br /><sub><b>Mariia</b></sub></a><br /><a href="https://github.com/DelineaXPM/dsv-gitlab/commits?author=mariiatuzovska" title="Code">💻</a></td>
      <td align="center"><a href="https://www.sheldonhull.com/"><img src="https://avatars.githubusercontent.com/u/3526320?v=4?s=100" width="100px;" alt="sheldonhull"/><br /><sub><b>sheldonhull</b></sub></a><br /><a href="https://github.com/DelineaXPM/dsv-gitlab/commits?author=sheldonhull" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/andrii-zakurenyi"><img src="https://avatars.githubusercontent.com/u/85106843?v=4?s=100" width="100px;" alt="andrii-zakurenyi"/><br /><sub><b>andrii-zakurenyi</b></sub></a><br /><a href="https://github.com/DelineaXPM/dsv-gitlab/commits?author=andrii-zakurenyi" title="Code">💻</a></td>
      <td align="center"><a href="https://github.com/gg-delinea"><img src="https://avatars.githubusercontent.com/u/99193946?v=4?s=100" width="100px;" alt="gg-delinea"/><br /><sub><b>gg-delinea</b></sub></a><br /><a href="#userTesting-gg-delinea" title="User Testing">📓</a></td>
    </tr>
  </tbody>
  <tfoot>

  </tfoot>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
