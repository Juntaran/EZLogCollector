/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 17:31
  */

package harvester

import "github.com/pkg/errors"

var (
	ErrFileTruncate = errors.New("detected lcFile being truncated")
	ErrRenamed      = errors.New("lcFile was renamed")
	ErrRemoved      = errors.New("lcFile was removed")
	ErrInactive     = errors.New("lcFile inactive")
	ErrClosed       = errors.New("reader closed")
)
