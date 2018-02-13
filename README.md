# Secret Secretary

Thanks for the help and name for: https://github.com/bonifaido

Project goal is to make secret distribution between namespaces in kubernetes.

Use cases:
  - Distribute cert (eg. letsencrypt generated with multiple SANs) from one namespace to many, if your goal is to use multiple ingeresses in multiple namespaces
  - If you have an image pull secret from your kubernetes provider, you can distribute that

How to:
  - Config should be in a config.yaml file mounted to /app/config.yaml
  - There is an example helm chat under examples
  - If you leave empty `targetnamespaces` value, then it will copy secrets to all namespaces
  - You can modify the 300 sec default cycle time by adding CYCLE environment variable of your choosing
