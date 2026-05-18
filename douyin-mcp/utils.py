import requests
import hashlib
import base64
import time
from urllib.parse import urlparse

# 模拟 User-Agent
UA = 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) EdgiOS/121.0.2277.107 Version/17.0 Mobile/15E148 Safari/604.1'

class DouyinUtils:
    def __init__(self):
        pass

    @staticmethod
    def get_ttwid():
        """获取 ttwid，增强请求稳定性"""
        url = 'https://ttwid.bytedance.com/ttwid/union/register/'
        data = '{"region":"cn","aid":1768,"needFid":false,"service":"www.ixigua.com","migrate_info":{"ticket":"","source":"node"},"cbUrlProtocol":"https","union":true}'
        try:
            res = requests.post(url=url, data=data, timeout=10)
            for i, j in res.cookies.items():
                if i == 'ttwid':
                    return j
        except Exception:
            pass
        return None

    @staticmethod
    def get_xbogus(payload, ua=UA, form=''):
        """生成 X-Bogus 签名"""
        short_str = "Dkdpgh4ZKsQB80/Mfvw36XI1R25-WUAlEi7NLboqYTOPuzmFjJnryx9HVGcaStCe="
        
        # arr1 generation logic
        salt_payload_bytes = hashlib.md5(hashlib.md5(payload.encode()).digest()).digest()
        salt_payload = [byte for byte in salt_payload_bytes]

        salt_form_bytes = hashlib.md5(hashlib.md5(form.encode()).digest()).digest()
        salt_form = [byte for byte in salt_form_bytes]

        # _0x30492c inline logic
        def _hash_xor(a, b):
            d = [i for i in range(256)]
            c = 0
            result = bytearray(len(b))
            for i in range(256):
                c = (c + d[i] + ord(a[i % len(a)])) % 256
                d[i], d[c] = d[c], d[i]
            t = 0
            c = 0
            for i in range(len(b)):
                t = (t + 1) % 256
                c = (c + d[t]) % 256
                d[t], d[c] = d[c], d[t]
                result[i] = ord(b[i]) ^ d[(d[t] + d[c]) % 256]
            return result

        ua_key = ['\u0000', '\u0001', '\u000e']
        salt_ua_bytes = hashlib.md5(base64.b64encode(_hash_xor(ua_key, ua))).digest()
        salt_ua = [byte for byte in salt_ua_bytes]

        timestamp = int(time.time())
        canvas = 1489154074

        arr1 = [64, 0, 1, 14, salt_payload[14], salt_payload[15], salt_form[14], salt_form[15], salt_ua[14], salt_ua[15], (timestamp >> 24) & 255, (timestamp >> 16) & 255, (timestamp >> 8) & 255, (timestamp >> 0) & 255, (canvas >> 24) & 255, (canvas >> 16) & 255, (canvas >> 8) & 255, (canvas >> 0) & 255, 64]
        for i in range(1, len(arr1) - 1):
            arr1[18] ^= arr1[i]

        arr2 = [arr1[0], arr1[2], arr1[4], arr1[6], arr1[8], arr1[10], arr1[12], arr1[14], arr1[16], arr1[18], arr1[1], arr1[3], arr1[5], arr1[7], arr1[9], arr1[11], arr1[13], arr1[15], arr1[17]]
        
        # Garbled string
        p = [arr2[0], arr2[10], arr2[1], arr2[11], arr2[2], arr2[12], arr2[3], arr2[13], arr2[4], arr2[14], arr2[5], arr2[15], arr2[6], arr2[16], arr2[7], arr2[17], arr2[8], arr2[18], arr2[9]]
        char_array = "".join([chr(i) for i in p])
        f_bytes = _hash_xor(['ÿ'], char_array)
        f = [2, 255] + [b for b in f_bytes]

        xbogus = ""
        for i in range(0, 21, 3):
            base_num = f[i + 2] | f[i + 1] << 8 | f[i] << 16
            xbogus += short_str[(base_num & 16515072) >> 18]
            xbogus += short_str[(base_num & 258048) >> 12]
            xbogus += short_str[(base_num & 4032) >> 6]
            xbogus += short_str[base_num & 63]
            
        return xbogus

    @staticmethod
    def sign_url(url, ua=UA):
        """为 URL 添加 X-Bogus 签名"""
        payload = urlparse(url).query
        xbogus = DouyinUtils.get_xbogus(payload, ua)
        return f"{url}&X-Bogus={xbogus}"
