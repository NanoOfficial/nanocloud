// @filename: config.go
// @author: Krisna Pranav
// @license: Copyright (c) 2023, Nano Blocks, Krisna Pranav

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Path                string `yaml:"path"`
	AppFile             string `yaml:"appFile"`
	WindowTitle         string `yaml:"windowTitle"`
	HWKey               bool   `yaml:"hardwareKey"`
	AppMode             string `yaml:"appMode"`
	AppName             string `yaml:"appName"`
	ScreenWidth         int    `yaml:"screenWidth"`
	ScreenHeight        int    `yaml:"screenHeight"`
	IsWindowMode        *bool  `yaml:"isWindowMode"`
	DiscoveryHost       string `yaml:"discoveryHost"`
	InstanceAddr        string `yaml:"instanceAddr"`
	HasChat             bool   `yaml:"hasChat"`
	PageTitle           string `yaml:"pageTitle"`
	StunTurn            string `yaml:"stunturn"`
	VideoCodec          string `yaml:"videoCodec"`
	IsVirtualized       bool   `yaml:"virtualize"`
	NAT1To1IP           string `yaml:"nat1to1ip"`
	DisableInterceptors bool   `yaml:"disableInterceptors"`
}

type AppDiscoveryMeta struct {
	ID           string `json:"id"`
	AppName      string `json:"app_name"`
	Addr         string `json:"addr"`
	AppMode      string `json:"app_mode"`
	HasChat      bool   `json:"has_chat"`
	PageTitle    string `json:"page_title"`
	ScreenWidth  int    `json:"screen_width"`
	ScreenHeight int    `json:"screen_height"`
}

func ReadConfig(path string) (Config, error) {
	cfgyml, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{}
	err = yaml.Unmarshal(cfgyml, &cfg)

	if cfg.AppName == "" {
		cfg.AppName = cfg.WindowTitle
	}
	if cfg.ScreenWidth == 0 {
		cfg.ScreenWidth = 800
	}
	if cfg.ScreenHeight == 0 {
		cfg.ScreenHeight = 600
	}
	if cfg.IsWindowMode == nil {
		boolTrue := true
		cfg.IsWindowMode = &boolTrue
	}
	if cfg.InstanceAddr == "" {
		ip, _ := getLocalIP()
		cfg.InstanceAddr = fmt.Sprintf("%s:%s", ip.String(), "8080")
	}
	return cfg, err
}

func getLocalIP() (net.IP, error) {
	tt, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, t := range tt {
		aa, err := t.Addrs()
		if err != nil {
			return nil, err
		}
		for _, a := range aa {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			v4 := ipnet.IP.To4()
			if v4 == nil || v4[0] == 127 {
				continue
			}
			return v4, nil
		}
	}
	return nil, errors.New("cannot find local IP address")
}
