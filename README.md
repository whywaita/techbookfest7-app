# techbookfest7-app

## ディレクトリ構成

```
techbookfest7-app
├── README.md
├── sample-app
│   ├── terraform
|   |   ├── README.md
|   |   ├── seeds.yaml
|   |   ├── modules
|   |   |   └── provider
|   |   |       └── main.tf
|   |   ├── resources
|   |   |   ├── common
|   |   |   |   └── api
|   |   |   |       └── main.tf
|   |   |   ├── prd
|   |   |   |   ├── init.tf
|   |   |   |   ├── routers.tf
|   |   |   |   ├── service_account.tf
|   |   |   |   ├── vpc.tf
|   |   |   |   ├── variables.tf
|   |   |   |   ├── provider.tf
|   |   |   |   └── gke.tf
|   |   |   └── dev
|   |   └── .terraform-version
│   ├── kustomize
|   |   ├── README.md
|   |   ├── seeds.yaml
│   │   ├── base
│   │   │   ├── kustomization.yaml
│   │   │   └── service.yaml
│   │   └── overlays
│   │       ├── prd
│   │       │   ├── kustomization.yaml
│   │       │   └── service.yaml
│   │       └── dev
│   │           ├── kustomization.yaml
│   │           └── service.yaml
└── cluster-generator
    ├── terraform
    |   ├── README.md
    │   └── dir
    ├── kustomize
    |   ├── README.md
    │   ├── base
    │   │   ├── kustomization.yaml
    │   │   └── service.yaml
    │   └── overlays
    │       └── prd
    │           ├── kustomization.yaml
    │           └── service.yaml
    └── apps
        ├── README.md
        ├── gateway-app
        |   ├── README.md
        │   ├── Dockerfile
        │   └── dir
        ├── kustomization-app
        |   ├── README.md
        │   ├── Dockerfile
        │   └── dir
        └── terraforming-app
            ├── README.md
            ├── Dockerfile
            └── dir
```
