package mongo

import (
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"

	nftls "github.com/netlify/netlify-commons/tls"
)

const (
	CollectionBlobs         = "blobs"
	CollectionResellers     = "resellers"
	CollectionUsers         = "users"
	CollectionSubscriptions = "bb_subscriptions"
	CollectionSites         = "projects"
)

type Config struct {
	TLS         *nftls.Config `mapstructure:"tls_conf"`
	DB          string        `mapstructure:"db"`
	Servers     []string      `mapstructure:"servers"`
	ConnTimeout int64         `mapstructure:"conn_timeout"`
}

func Connect(config *Config, log *logrus.Entry) (*mgo.Database, error) {
	info := &mgo.DialInfo{
		Addrs:   config.Servers,
		Timeout: time.Second * time.Duration(config.ConnTimeout),
	}

	if config.TLS != nil {
		tlsLog := log.WithFields(logrus.Fields{
			"cert_file": config.TLS.CertFile,
			"key_file":  config.TLS.KeyFile,
			"ca_files":  strings.Join(config.TLS.CAFiles, ","),
		})

		tlsLog.Debug("Using TLS config")
		tlsConfig, err := config.TLS.TLSConfig()
		if err != nil {
			return nil, err
		}

		info.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), tlsConfig)
		}
	} else {
		log.Debug("Skipping TLS config")
	}

	log.WithField("servers", strings.Join(info.Addrs, ",")).Debug("Dialing database")
	sess, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, err
	}

	log.WithField("db", config.DB).Debugf("Got session, Using database %s", config.DB)
	return sess.DB(config.DB), nil
}
