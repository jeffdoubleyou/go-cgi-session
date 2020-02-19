package sessionid

import(
    "time"
    "crypto/md5"
    "math/rand"
    "os"
    "fmt"
)

type Md5Id struct {

}

func Md5() *Md5Id {
    return &Md5Id{}
}

func (m *Md5Id) Generate() string {
    now := time.Now()
    rand.Seed(now.UnixNano())
    s := fmt.Sprintf("%d%d%d", os.Getpid(), now.Unix(), rand.Intn(int(now.Unix())))
    return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
