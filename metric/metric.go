package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/siaoynli/go-project-simple/config"
	"github.com/siaoynli/pkg/prome"
)

var ProductSearch = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:        "shop_product_search",
		Help:        "histogram for product search",
		Buckets:     prome.DefaultBuckets,
		ConstLabels: prometheus.Labels{"machine": prome.GetHostName(), "app": config.AppName},
	},
	[]string{"cluster"},
)

var OrderSearch = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:        "shop_order_search",
		Help:        "histogram for order search",
		Buckets:     prome.DefaultBuckets,
		ConstLabels: prometheus.Labels{"machine": prome.GetHostName(), "app": config.AppName},
	},
	[]string{"cluster"},
)
