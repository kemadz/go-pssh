package cmd

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/xuchenCN/go-pssh/yaml"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	hostFile string
	hostList string
	configFile string

	Hosts []string  `json:"hosts"`
	Port int 		`json:port`
	User string 	`json:user`
	Password string	`json:"password"`
	Cmd string		`json:"cmd"`
}

func (c *config) addFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&c.configFile,"config","y","","config file format in yaml or json")
	fs.StringVarP(&c.hostFile,"file","f","","file path of hosts")
	fs.StringVarP(&c.hostList,"list","l","","hosts:ip1,ip2")
	fs.IntVarP(&c.Port,"port","p",22,"port of ssh connect to")
	fs.StringVarP(&c.User,"user","u","root","user")
	fs.StringVarP(&c.Password,"password","P","","password")
	fs.StringVarP(&c.Cmd,"cmd","c","","command")
}

func (c *config) validate() error {

	if len(c.hostFile) <= 0 && len(c.hostList) <= 0 && len(c.configFile) <= 0{
		return fmt.Errorf("need file of host or hosts list or config file")
	}

	if len(c.configFile) > 0 {
		if abs ,err := filepath.Abs(c.configFile) ; err != nil {
			return nil
		} else {
			c.configFile = abs
			log.Infof("convert path to %s" , c.configFile)
		}

		c.loadConfigFile()
	}

	if len(c.Cmd) <= 0 {
		return fmt.Errorf("where is your command")
	}

	if len(c.hostFile) > 0 {
		if abs ,err := filepath.Abs(c.hostFile) ; err != nil {
			return nil
		} else {
			c.hostFile = abs
			log.Infof("convert path to %s" , c.hostFile)
		}
	}

	return nil
}

func (c *config) loadConfigFile() error {
	if len(c.configFile) <= 0 {
		return fmt.Errorf("use -cfg to locate config-file")
	}

	if !filepath.IsAbs(c.configFile) {
		return fmt.Errorf("expect the abosulte path of config-file")
	}

	if cfgFile ,err := os.Open(c.configFile); err == nil {
		yamlToJsonDecoder := yaml.NewYAMLToJSONDecoder(cfgFile)
		return yamlToJsonDecoder.Decode(&c)
	} else {
		return err
	}

	return nil
}

func (c *config) listHosts() ([]string, error) {
	result := make(map[string]interface{})
	if len(c.hostFile) > 0 {
		if file, err := os.Open(c.hostFile); err == nil {
			defer file.Close()
			fr := bufio.NewReader(file)
			for {
				b, _, err := fr.ReadLine()
				if err == io.EOF {
					break;
				}
				line := strings.TrimSpace(string(b))
				if ip := net.ParseIP(line); ip == nil {
					log.Error("%s is not valid ip addr ignore it")
					continue
				}

				result[line] = nil
			}
		} else {
			return nil,err
		}
	}

	if len(c.hostList) > 0 {
		list := strings.Split(c.hostList,",")
		for _,host := range list {
			host = strings.TrimSpace(host)
			if ip := net.ParseIP(host); ip == nil {
				log.Error("%s is not valid ip addr ignore it")
				continue
			}

			result[host] = nil
		}
	}

	if len(c.Hosts) > 0 {
		for _, host := range c.Hosts {
			result[host] = nil
		}
	}

	keys := make([]string, 0, len(result))
	for k,_ := range result {
		keys = append(keys,k)
	}

	return keys,nil
}