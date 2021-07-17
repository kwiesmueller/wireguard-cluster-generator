package main

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

func generateKeyPair() (string, string, error) {
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	return privateKey.String(), privateKey.PublicKey().String(), nil
}
