package main

import (
	"github.com/vugu/vgrouter"
	"github.com/vugu/vugu"
)

// https://www.vugu.org/doc/routing
func vuguSetup(buildEnv *vugu.BuildEnv, eventEnv vugu.EventEnv) vugu.Builder {
	router := vgrouter.New(eventEnv)
	buildEnv.SetWireFunc(func(b vugu.Builder) {
		if c, ok := b.(vgrouter.NavigatorSetter); ok {
			c.NavigatorSet(router)
		}
	})

	//

	root := &Root{}
	buildEnv.WireComponent(root)

	router.MustAddRouteExact("/",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &Upload{root: root}
		}),
	)

	router.MustAddRouteExact("/download/:artifact",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &Download{root: root, RouteMatch: rm}
		}),
	)

	router.MustAddRouteExact("/download",
		vgrouter.RouteHandlerFunc(func(rm *vgrouter.RouteMatch) {
			root.Body = &Download{root: root, RouteMatch: rm, NoArtifact: true}
		}),
	)

	// router.SetNotFound(vgrouter.RouteHandlerFunc(
	// 	func(rm *vgrouter.RouteMatch) {
	// 		root.Body = &PageNotFound{}
	// 	}))

	// router.SetUseFragment(true)

	err := router.ListenForPopState()
	if err != nil {
		panic(err)
	}

	err = router.Pull()
	if err != nil {
		panic(err)
	}
	return root
}
