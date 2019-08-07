# sample-app/terraform

## 事前準備
- mac
- cloud sdkのインストール
- terraformのインストール
- GCPプロジェクトの作成
- Google Compute Engine APIの有効化
- Service Accountの作成

## cluster-generatorが触る箇所
- tfvars
- resources/dev/

## ディレクトリ構成

```
terraform
├── README.md
├── modules
└── resources
    ├── common
    ├── dev
    |   ├── credentials
    |   ├── gke.tf
    |   ├── init.tf
    |   ├── provider.tf
    |   ├── routers.tf
    |   ├── service_account.tf
    |   ├── variables.tf
    |   ├── vpc.tf
    |   └── .terraform-version
    └── prd
        ├── credentials
        ├── gke.tf
        ├── init.tf
        ├── provider.tf
        ├── routers.tf
        ├── service_account.tf
        ├── variables.tf
        ├── vpc.tf
        └── .terraform-version
```