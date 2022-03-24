package routing

import (
	"errors"
	"path"
	"sort"
	"strconv"
	"strings"
)

type (

	// UrlResolver is a URL resolver utility that stores the handler information registered by UrlFor.
	UrlResolver interface {
		// Get is a resolver function that takes a name and parameters and returns a URL. If the URL is not found, it panics.
		Get(urlName string, params ...string) string
		// Reverse is a resolver function that takes a name and parameters and returns a URL.
		Reverse(urlName string, params ...string) (string, error)
		// ReverseWithParams is a resolver function that takes a name and parameters and returns a URL.
		ReverseWithParams(urlName string, params []string) (string, error)
		// MustReverse is a resolver function that takes a name and parameters and returns a URL. If the URL is not found, it panics.
		MustReverse(urlName string, params ...string) string
		// MustReverseWithParams is a resolver function that takes a name and parameters and returns a URL. If the URL is not found, it panics.
		MustReverseWithParams(urlName string, params []string) string
	}

	// UrlFor is a reverse-routing utility that stores the handler information.
	UrlFor interface {
		UrlResolver
		// Add registers name, parameter, and URL for UrlResolver.
		// If a duplicate name exists, an error is returned instead of registering.
		Add(urlName, urlAddr string, params ...string) (string, error)
		// MustAdd registers name, parameter, and URL for UrlResolver.
		// If a duplicate name exists, it panics.
		MustAdd(urlName, urlAddr string, params ...string) string
		// AddGr registers name, parameter, and URL for UrlResolver with nested group infos.
		// If a duplicate name exists, an error is returned instead of registering.
		AddGr(urlName, urlAddr string, groupNames, groupAddrs []string, params ...string) (string, error)
		// MustAddGr registers name, parameter, and URL for UrlResolver with nested group infos.
		// If a duplicate name exists, it panics.
		MustAddGr(urlName, urlAddr string, groupNames, groupAddrs []string, params ...string) string
		// Clear clears all registered reverse-routing infos.
		Clear()
		// String returns summarized info of registered reverse-routing infos.
		String() string
		// ToResolver returns a UrlResolver.
		ToResolver() UrlResolver
	}
	routerFragment struct {
		url    string
		params []string
	}

	reverseRouter map[string]routerFragment

	reverseRouteResolver func(urlName string, params []string) (string, error)
)

// NewUrlFor returns a UrlFor.
func NewUrlFor() UrlFor {
	router := make(reverseRouter)
	return &router
}

func (rr reverseRouteResolver) Get(urlName string, params ...string) string {
	res, err := rr(urlName, params)
	if err != nil {
		panic(err)
	}
	return res
}

func (rr reverseRouteResolver) Reverse(urlName string, params ...string) (string, error) {
	return rr(urlName, params)
}

func (rr reverseRouteResolver) ReverseWithParams(urlName string, params []string) (string, error) {
	return rr(urlName, params)
}

func (rr reverseRouteResolver) MustReverse(urlName string, params ...string) string {
	res, err := rr(urlName, params)
	if err != nil {
		panic(err)
	}
	return res
}

func (rr reverseRouteResolver) MustReverseWithParams(urlName string, params []string) string {
	res, err := rr(urlName, params)
	if err != nil {
		panic(err)
	}
	return res
}

func (us *reverseRouter) ToResolver() UrlResolver {
	return reverseRouteResolver(us.ReverseWithParams)
}

func (us *reverseRouter) MustReverse(urlName string, params ...string) string {
	res, err := us.ReverseWithParams(urlName, params)
	if err != nil {
		panic(err)
	}
	return res
}

func (us *reverseRouter) MustReverseWithParams(urlName string, params []string) string {
	res, err := us.ReverseWithParams(urlName, params)
	if err != nil {
		panic(err)
	}
	return res
}

func (us *reverseRouter) MustAdd(urlName, urlAddr string, params ...string) string {
	addr, err := us.addInternal(urlName, urlAddr, nil, nil, params)
	if err != nil {
		panic(err)
	}
	return addr
}

func (us *reverseRouter) Add(urlName, urlAddr string, params ...string) (string, error) {
	return us.addInternal(urlName, urlAddr, nil, nil, params)
}

func (us *reverseRouter) MustAddGr(urlName, urlAddr string, groupNames, groupAddrs []string, params ...string) string {
	addr, err := us.addInternal(urlName, urlAddr, groupNames, groupAddrs, params)
	if err != nil {
		panic(err)
	}
	return addr
}

func (us *reverseRouter) AddGr(urlName, urlAddr string, groupNames, groupAddrs []string, params ...string) (string, error) {
	return us.addInternal(urlName, urlAddr, groupNames, groupAddrs, params)
}

func (us *reverseRouter) Reverse(urlName string, params ...string) (string, error) {
	return us.ReverseWithParams(urlName, params)
}

func (us reverseRouter) addInternal(urlName, urlAddr string, groupNames, groupAddrs, params []string) (string, error) {
	if _, ok := us[urlName]; ok {
		return "", errors.New("Url already exists. Try to use .Get() method.")
	}
	routeName := strings.Join(append(groupNames, urlName), ".")
	addr := path.Join(append(groupAddrs, urlAddr)...)

	tmpUrl := routerFragment{addr, params}
	us[routeName] = tmpUrl
	return addr, nil
}

func (us reverseRouter) Clear() {
	for k := range us {
		delete(us, k)
	}
}

func (us *reverseRouter) Get(urlName string, params ...string) string {
	url, err := us.ReverseWithParams(urlName, params)
	if err != nil {
		panic(err)
	}
	return url
}

func (us reverseRouter) ReverseWithParams(urlName string, params []string) (string, error) {
	if len(params) != len(us[urlName].params) {
		return "", errors.New("Bad Url Reverse: mismatch params for URL: " + urlName)
	}
	res := us[urlName].url
	for i, val := range params {
		res = strings.Replace(res, us[urlName].params[i], val, 1)
	}
	return res, nil
}

func (us reverseRouter) String() (ret string) {
	var (
		numOfRoutes = len(us)
		builder     strings.Builder
		needSort    bool
	)
	defer func() {
		ret = builder.String()
	}()

	builder.WriteString(strconv.FormatInt(int64(numOfRoutes), 10))
	builder.WriteByte(' ')
	switch numOfRoutes {
	case 0:
		builder.WriteString("route")
		return
	case 1:
		builder.WriteString("route:\n")
	default:
		builder.WriteString("routes:\n")
		needSort = true
	}

	fragmentStringer := func(builder *strings.Builder, idx int, key string, value routerFragment) {
		builder.WriteByte('\t')
		builder.WriteString(key)
		builder.WriteByte('(')
		builder.WriteString(value.url)
		builder.WriteString(") [")
		numOfParams := len(value.params)
		for pathIdx := 0; pathIdx < numOfParams; pathIdx++ {
			builder.WriteString(value.params[pathIdx])
			if pathIdx < numOfParams-1 {
				builder.WriteByte(',')
			}
		}
		builder.WriteByte(']')
		if idx < numOfRoutes-1 {
			builder.WriteByte('\n')
		}
	}
	if needSort {
		routeKeys := make([]string, 0, numOfRoutes)
		for key := range us {
			routeKeys = append(routeKeys, key)
		}

		sort.SliceStable(routeKeys, func(i, j int) bool {
			return len(us[routeKeys[i]].url) < len(us[routeKeys[j]].url)
		})
		for i, key := range routeKeys {
			fragmentStringer(&builder, i, key, us[key])
		}
	} else {
		i := 0
		for key, value := range us {
			fragmentStringer(&builder, i, key, value)
			i++
		}
	}
	return
}

func (us reverseRouter) getParamName(urlName string, num int) string {
	return us[urlName].params[num]
}
