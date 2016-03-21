package main

const (
	SYS_CONF_LANG        = GLISP_SYMBOL_TYPE_LANG
	SYS_CONF_TIMEZONE    = GLISP_SYMBOL_TYPE_TIMEZONE
	SYS_CONF_NETWORK     = GLISP_SYMBOL_TYPE_NETWORK
	SYS_CONF_APP_VERSION = GLISP_SYMBOL_TYPE_APP_VERSION
	SYS_CONF_OS_TYPE     = GLISP_SYMBOL_TYPE_OS_TYPE
	SYS_CONF_OS_VERSION  = GLISP_SYMBOL_TYPE_OS_VERSION
	SYS_CONF_IP          = GLISP_SYMBOL_TYPE_IP
	SYS_CONF_DEVICE_ID   = GLISP_SYMBOL_TYPE_DEVICE_ID
	SYS_CONF_APP_KEY     = GLISP_SYMBOL_TYPE_DEVICE_ID
)

func isSysConfType(appKey string) bool {
	return appKey == SYS_CONF_LANG ||
		appKey == SYS_CONF_TIMEZONE ||
		appKey == SYS_CONF_NETWORK ||
		appKey == SYS_CONF_APP_VERSION ||
		appKey == SYS_CONF_OS_TYPE ||
		appKey == SYS_CONF_IP ||
		appKey == SYS_CONF_DEVICE_ID ||
		appKey == SYS_CONF_APP_KEY ||
		appKey == SYS_CONF_OS_VERSION

}

func uniformClientParams(cdata *ClientData) *ClientData {
	c := *cdata

	configs := getAppMatchConfWithKey(SYS_CONF_APP_KEY, cdata, cdata.AppKey)
	if len(configs) > 0 {
		newV := configs[cdata.AppKey]
		if v, ok := newV.(string); ok {
			c.AppKey = v
		}
	}

	configs = getAppMatchConfWithKey(SYS_CONF_LANG, cdata, cdata.Lang)
	if len(configs) > 0 {
		newV := configs[cdata.Lang]
		if v, ok := newV.(string); ok {
			c.Lang = v
		}
	}

	configs = getAppMatchConfWithKey(SYS_CONF_IP, cdata, cdata.Ip)
	if len(configs) > 0 {
		newV := configs[cdata.Ip]
		if v, ok := newV.(string); ok {
			c.Ip = v
		}
	}

	configs = getAppMatchConfWithKey(SYS_CONF_DEVICE_ID, cdata, cdata.DeviceId)
	if len(configs) > 0 {
		newV := configs[cdata.DeviceId]
		if v, ok := newV.(string); ok {
			c.DeviceId = v
		}
	}

	configs = getAppMatchConfWithKey(SYS_CONF_TIMEZONE, cdata, cdata.TimeZone)
	if len(configs) > 0 {
		newV := configs[cdata.TimeZone]
		if v, ok := newV.(string); ok {
			c.TimeZone = v
		}
	}

	configs = getAppMatchConfWithKey(SYS_CONF_NETWORK, cdata, cdata.NetWork)
	if len(configs) > 0 {
		newV := configs[cdata.NetWork]
		if v, ok := newV.(string); ok {
			c.NetWork = v
		}
	}

	configs = getAppMatchConfWithKey(SYS_CONF_APP_VERSION, cdata, cdata.AppVersion)
	if len(configs) > 0 {
		newV := configs[cdata.AppVersion]
		if v, ok := newV.(string); ok {
			c.AppVersion = v
		}
	}

	configs = getAppMatchConfWithKey(SYS_CONF_OS_TYPE, cdata, cdata.OSType)
	if len(configs) > 0 {
		newV := configs[cdata.OSType]
		if v, ok := newV.(string); ok {
			c.OSType = v
		}
	}

	configs = getAppMatchConfWithKey(SYS_CONF_OS_VERSION, cdata, cdata.OSVersion)
	if len(configs) > 0 {
		newV := configs[cdata.OSVersion]
		if v, ok := newV.(string); ok {
			c.OSVersion = v
		}
	}

	return &c
}
