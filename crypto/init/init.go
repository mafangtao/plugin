package init

//系统级别的签名定义
//为了安全考虑，默认情况下，我们希望只定义合约内部的签名，系统级别的签名对所有的合约都有效

import (
	_ "gitlab.33.cn/chain33/plugin/crypto/ecdsa"
	_ "gitlab.33.cn/chain33/plugin/crypto/sm2"
)
