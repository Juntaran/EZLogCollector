/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 17:46
  */

package harvester

import (
	"expvar"
)

var (
	harvesterStarted   = expvar.NewInt("filebeat.harvester.started")
	harvesterClosed    = expvar.NewInt("filebeat.harvester.closed")
	harvesterRunning   = expvar.NewInt("filebeat.harvester.running")
	harvesterOpenFiles = expvar.NewInt("filebeat.harvester.open_files")
	filesTruncated     = expvar.NewInt("filebeat.harvester.files.truncated")
)

