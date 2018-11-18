/** 
  * Author: Juntaran 
  * Email:  Jacinthmail@gmail.com 
  * Date:   2018/11/18 01:35
  */

package streambuf

import "errors"


var ErrOperationNotAllowed = errors.New("Operation not allowed")

var ErrOutOfRange = errors.New("Data access out of range")

var ErrNoMoreBytes = errors.New("No more bytes")

var ErrUnexpectedEOB = errors.New("unexpected end of buffer")

var ErrExpectedByteSequenceMismatch = errors.New("expected byte sequence did not match")