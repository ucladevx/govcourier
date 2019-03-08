package main

import (
	"fmt"
	"github.com/hackform/governor"
	"github.com/hackform/governor/service/barcode"
	"github.com/hackform/governor/service/cache"
	"github.com/hackform/governor/service/cache/conf"
	"github.com/hackform/governor/service/cachecontrol"
	"github.com/hackform/governor/service/conf"
	"github.com/hackform/governor/service/conf/model"
	"github.com/hackform/governor/service/courier"
	"github.com/hackform/governor/service/courier/model"
	"github.com/hackform/governor/service/db"
	"github.com/hackform/governor/service/db/conf"
	"github.com/hackform/governor/service/mail"
	"github.com/hackform/governor/service/mail/conf"
	"github.com/hackform/governor/service/msgqueue"
	"github.com/hackform/governor/service/msgqueue/conf"
	"github.com/hackform/governor/service/objstore"
	"github.com/hackform/governor/service/objstore/conf"
	"github.com/hackform/governor/service/template"
	"github.com/hackform/governor/service/user"
	"github.com/hackform/governor/service/user/conf"
	"github.com/hackform/governor/service/user/gate"
	"github.com/hackform/governor/service/user/gate/conf"
	"github.com/hackform/governor/service/user/model"
	"github.com/hackform/governor/service/user/role/model"
)

var (
	// GitHash is a build hash to be passed in at compile time
	GitHash string
)

func main() {
	config, err := governor.NewConfig("config", GitHash)
	governor.Must(err)

	fmt.Println("created new config")
	fmt.Println("loading config defaults:")

	governor.Must(dbconf.Conf(&config))
	fmt.Println("- db")

	governor.Must(cacheconf.Conf(&config))
	fmt.Println("- cache")

	governor.Must(objstoreconf.Conf(&config))
	fmt.Println("- objstore")

	governor.Must(msgqueueconf.Conf(&config))
	fmt.Println("- msgqueue")

	governor.Must(mailconf.Conf(&config))
	fmt.Println("- mail")

	governor.Must(gateconf.Conf(&config))
	fmt.Println("- gate")

	governor.Must(userconf.Conf(&config))
	fmt.Println("- user")

	governor.Must(config.Init())
	fmt.Println("config loaded")

	l := governor.NewLogger(config)

	g, err := governor.New(config, l)
	governor.Must(err)

	dbService, err := db.New(config, l)
	governor.Must(err)

	cacheService, err := cache.New(config, l)
	governor.Must(err)

	objstoreService, err := objstore.New(config, l)
	governor.Must(err)

	queueService, err := msgqueue.New(config, l)
	governor.Must(err)

	templateService, err := template.New(config, l)
	governor.Must(err)

	mailService, err := mail.New(config, l, templateService, queueService)
	governor.Must(err)

	gateService := gate.New(config, l)

	barcodeService := barcode.New(config, l)

	cacheControlService := cachecontrol.New(config, l)

	confRepo := confmodel.New(config, l, dbService)
	confService := conf.New(config, l, confRepo)

	roleModelService := rolemodel.New(config, l, dbService)
	userModelService := usermodel.New(config, l, dbService, roleModelService)
	userService := user.New(config, l, userModelService, roleModelService, cacheService, mailService, gateService, cacheControlService)

	courierModelService := couriermodel.New(config, l, dbService)
	courierService, err := courier.New(config, l, courierModelService, objstoreService, barcodeService, cacheService, gateService, cacheControlService)
	governor.Must(err)

	governor.Must(g.MountRoute("/null/database", dbService))
	governor.Must(g.MountRoute("/null/cache", cacheService))
	governor.Must(g.MountRoute("/null/objstore", objstoreService))
	governor.Must(g.MountRoute("/null/msgqueue", queueService))
	governor.Must(g.MountRoute("/null/mail", mailService))
	governor.Must(g.MountRoute("/conf", confService))
	governor.Must(g.MountRoute("/u", userService))
	governor.Must(g.MountRoute("/courier", courierService))

	governor.Must(g.Start())
}
