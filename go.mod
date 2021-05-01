module github.com/lavelle96/seldon-deployment

go 1.16

require (
	github.com/seldonio/seldon-core/operator v0.0.0-20210305115125-18a2c688413c
	k8s.io/api v0.18.8 // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.4 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
