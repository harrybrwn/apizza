// Package dawg (Dominos API Wrapper for Go) is a package that allows a go programmer
// to interact with the dominos web api.
//
// The two main entry points in the package are the SignIn and NearestStore functions.
//
// SignIn is used if you have an account with dominos.
// 	user, err := dawg.SignIn(username, password)
//	if err != nil {
// 		// handle error
// 	}
//
// NearestStore should be used if you just want to make a one time order.
// 	store, err := dawg.NearestStore(&address, dawg.Delivery)
// 	if err != nil {
// 		// handle error
// 	}
//
// To order anything from dominos you need to find a store, create an order,
// then send that order.
package dawg

const (
	// DefaultLang is the package language variable
	DefaultLang = "en"

	orderHost = "order.dominos.com"
)
