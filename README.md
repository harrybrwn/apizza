# apizza
A cli for ordering domios pizza.

![Build Status](https://travis-ci.org/harrybrwn/apizza.svg?branch=master)

#### Installation
```
go install -u github.com/harrybrwn/apizza
```

#### Setup
```
apizza config set name="Harry Brown"
apizza config set email="harrybrown98@gmail.com"
```

#### The Domios API Wrapper for Go (dawg)
```
go get github.com/harrybrwn/apizza/dawg
```

###### Demo
```go
import (
	"github.com/harrybrwn/apizza/dawg"
)

var addr = &dawg.Address{
	Street: "1600 Pennsylvania Ave.",
	City: "Washington",
	State: "DC",
	Zip: "20500",
	AddrType: "House",
}

func main() {
	store, err := dawg.NearestStore(addr, "Delivery")
	if err != nil {
		panic(err)
	}
	pizza, err := store.GetProduct("16SCREEN")
}
```

happy coding :)
![hahaha](https://external-preview.redd.it/L5a31wsfcT9TcNcvOF3HTOFkXxnKjA7OopCakXxScDg.png?auto=webp&s=bbb10ca8d08363bc2d94996a77619d8bf60c24e8)
