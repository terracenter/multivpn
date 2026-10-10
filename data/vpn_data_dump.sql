PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE vpn_connections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    empresa TEXT,
    is_salida INTEGER DEFAULT 0
);
INSERT INTO vpn_connections VALUES(1,'FibexValencia','vpn','fibex',0);
INSERT INTO vpn_connections VALUES(2,'le','wireguard','minipc',1);
CREATE TABLE vpn_routes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vpn_id INTEGER NOT NULL,
    route TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY(vpn_id) REFERENCES vpn_connections(id) ON DELETE CASCADE,
    UNIQUE(vpn_id, route)
);
INSERT INTO vpn_routes VALUES(1,1,'192.168.226.16/28','PVE-VALENCIA-NOC');
INSERT INTO vpn_routes VALUES(2,1,'192.168.235.16/28','PVE-LA-MARACAY-NOC-01');
INSERT INTO vpn_routes VALUES(3,1,'172.16.9.0/24','VLAN 9 EL PARAISO');
INSERT INTO vpn_routes VALUES(4,1,'172.16.11.0/24','VLAN 11 EL PARAISO');
INSERT INTO vpn_routes VALUES(5,1,'172.16.13.0/24','VLAN 13 EL PARAISO');
INSERT INTO vpn_routes VALUES(6,1,'172.16.15.0/24','VLAN 15 EL PARAISO');
INSERT INTO vpn_routes VALUES(7,1,'172.16.16.0/24','VLAN 16 EL PARAISO');
INSERT INTO vpn_routes VALUES(8,1,'172.16.17.0/24','VLAN 17 EL PARAISO');
INSERT INTO vpn_routes VALUES(9,1,'172.16.20.0/24','VLAN 20 EL PARAISO');
INSERT INTO vpn_routes VALUES(10,1,'192.168.16.0/24','CARACAS EL PARAISO LAN');
INSERT INTO vpn_routes VALUES(11,1,'192.168.17.0/24','CARACAS EL PARAISO LAN');
INSERT INTO vpn_routes VALUES(12,1,'192.168.19.0/24','CARACAS EL PARAISO LAN');
INSERT INTO vpn_routes VALUES(13,1,'192.168.20.0/24','CARACAS EL PARAISO LAN');
INSERT INTO vpn_routes VALUES(14,1,'192.168.65.0/24','CARACAS LA URBINA LAN');
INSERT INTO vpn_routes VALUES(16,1,'10.235.245.0/24','LAN VPN EL PARAISO');
INSERT INTO vpn_routes VALUES(17,1,'10.254.254.0/24','LAN VPN DISLNET');
INSERT INTO vpn_routes VALUES(18,1,'10.10.20.48/29','VALENCIA DISTELNET PROFIT');
INSERT INTO vpn_routes VALUES(19,1,'10.10.20.84/28','VALENCIA DISTELNET IPTV');
INSERT INTO vpn_routes VALUES(20,1,'10.10.20.0/24','LAN Idrac Valencia');
INSERT INTO vpn_routes VALUES(21,1,'10.10.30.16/28','LAN SERVIDORES TORRE FIBEX');
INSERT INTO vpn_routes VALUES(22,1,'172.17.255.0/24','Distelnet Valencia');
INSERT INTO vpn_routes VALUES(23,1,'192.168.233.0/28','Distelnet Valencia');
INSERT INTO vpn_routes VALUES(24,1,'192.168.226.0/24','LAN SERVIDORES SUPER MICRO');
INSERT INTO vpn_routes VALUES(25,1,'192.168.245.0/24','VALENCIA TORRE FIBEX PISO-1');
INSERT INTO vpn_routes VALUES(26,1,'192.168.246.0/24','VALENCIA TORRE FIBEX PISO-2');
INSERT INTO vpn_routes VALUES(27,1,'192.168.247.0/24','VALENCIA TORRE FIBEX PISO-3');
INSERT INTO vpn_routes VALUES(28,1,'192.168.248.0/24','VALENCIA TORRE FIBEX PISO-4');
INSERT INTO vpn_routes VALUES(29,1,'190.103.30.10/32','pbx-sip-trunk-digitel');
INSERT INTO vpn_routes VALUES(30,1,'190.103.28.122/32','nginx-proxy-reverse');
INSERT INTO vpn_routes VALUES(31,1,'190.103.30.154/32','pbx.fibextelecom.net');
INSERT INTO vpn_routes VALUES(32,1,'192.168.227.0/24','IPTV Fibex el Paraiso');
INSERT INTO vpn_routes VALUES(33,1,'190.89.28.86/32','Astra PERU IPTV');
INSERT INTO vpn_routes VALUES(34,1,'148.113.162.189/32','PVE-OZMAP-OVH');
INSERT INTO vpn_routes VALUES(35,1,'190.103.29.34/32','DNS2-FIBEXTELECOM');
INSERT INTO vpn_routes VALUES(36,1,'190.103.30.142/32','DNS1-FIBEXTELECOM');
INSERT INTO vpn_routes VALUES(37,1,'192.168.233.0/24','Red desconocida');
INSERT INTO vpn_routes VALUES(39,1,'190.103.30.0/24','SIP-TRUNK-DIGITEL');
INSERT INTO vpn_routes VALUES(40,1,'10.255.204.48/28','LAN RD PVE-RD-IPTV-01');
INSERT INTO vpn_routes VALUES(41,1,'10.255.204.0/28','REPUBLICA DOMINICANA');
INSERT INTO vpn_routes VALUES(43,1,'192.168.168.16/28','LAN SERVIDORES PVE-CANADA RB-MK-FIBEXPLAY');
INSERT INTO vpn_routes VALUES(44,1,'149.56.31.17/32','PVE OVH Jose Manuel');
INSERT INTO vpn_routes VALUES(45,1,'142.44.173.240/30','Soyustart Canada');
INSERT INTO vpn_routes VALUES(46,1,'51.161.36.205/32','GLPI-WEB Soyiutart Canadas');
INSERT INTO vpn_routes VALUES(47,1,'51.79.17.76','PVE-OVH-CANADA-02');
INSERT INTO vpn_routes VALUES(48,1,'15.235.37.97','MK-RO-PVE-OVH-CANADA-IT-01');
INSERT INTO vpn_routes VALUES(50,1,'15.235.32.252','NGINX-PROXY-REVERSE');
INSERT INTO vpn_routes VALUES(51,1,'200.35.151.10','IPTV-PENTA-RD');
INSERT INTO vpn_routes VALUES(52,2,'10.6.0.0/24','Red interna PiVPN');
INSERT INTO vpn_routes VALUES(53,2,'138.1.0.0/16','Oracle Cloud Broad');
INSERT INTO vpn_routes VALUES(54,2,'104.64.0.0/10','Akamai Range 1');
INSERT INTO vpn_routes VALUES(55,2,'23.0.0.0/8','Akamai Broad Range');
INSERT INTO vpn_routes VALUES(56,2,'160.79.104.0/23','Claude AI');
INSERT INTO vpn_routes VALUES(57,2,'209.249.57.0/24','Claude AI');
INSERT INTO vpn_routes VALUES(58,2,'34.162.46.92/32','Claude AI');
INSERT INTO vpn_routes VALUES(59,2,'104.18.27.175/32','Cloudflare');
INSERT INTO vpn_routes VALUES(60,2,'8.216.131.0/24','Z.AI Network');
INSERT INTO vpn_routes VALUES(61,2,'104.244.40.0/21','104.244.40.0/21');
INSERT INTO vpn_routes VALUES(62,2,'199.96.56.0/21','199.96.56.0/21');
INSERT INTO vpn_routes VALUES(63,2,'192.133.76.0/22','192.133.76.0/22');
INSERT INTO vpn_routes VALUES(64,2,'199.16.156.0/22','199.16.156.0/22');
INSERT INTO vpn_routes VALUES(65,2,'199.59.148.0/22','199.59.148.0/22');
INSERT INTO vpn_routes VALUES(66,2,'8.25.194.0/23','8.25.194.0/23');
CREATE TABLE vpn_domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vpn_id INTEGER NOT NULL,
    domain TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY(vpn_id) REFERENCES vpn_connections(id) ON DELETE CASCADE,
    UNIQUE(vpn_id, domain)
);
INSERT INTO vpn_domains VALUES(1,2,'claude.ai','claude.ai');
INSERT INTO vpn_domains VALUES(2,2,'claude.com','claude.com');
INSERT INTO vpn_domains VALUES(3,2,'www.claude.ai','www.claude.ai');
INSERT INTO vpn_domains VALUES(4,2,'chatgpt.com','chatgpt.com');
INSERT INTO vpn_domains VALUES(5,2,'auth.openai.com','auth.openai.com');
INSERT INTO vpn_domains VALUES(6,2,'openai.com','openai.com');
INSERT INTO vpn_domains VALUES(7,2,'virtualbox.org','virtualbox.org');
INSERT INTO vpn_domains VALUES(8,2,'spotifycdn.com','spotifycdn.com');
INSERT INTO vpn_domains VALUES(9,2,'cdn.oaistatic.com','cdn.oaistatic.com');
INSERT INTO vpn_domains VALUES(10,2,'www.hostinger.com','www.hostinger.com');
INSERT INTO vpn_domains VALUES(11,2,'hostinger.com','hostinger.com');
INSERT INTO vpn_domains VALUES(12,2,'www.cloudflare.com/','www.cloudflare.com/');
INSERT INTO vpn_domains VALUES(13,2,'copilot.microsoft.com','copilot.microsoft.com');
INSERT INTO vpn_domains VALUES(14,2,'api-dashboard.search.brave.com','api-dashboard.search.brave.com');
INSERT INTO vpn_domains VALUES(15,2,'iforgot.apple.com','iforgot.apple.com');
INSERT INTO vpn_domains VALUES(16,2,'apple.com','apple.com');
INSERT INTO vpn_domains VALUES(17,2,'manychat.com','manychat.com');
INSERT INTO vpn_domains VALUES(18,2,'x.com','x.com');
INSERT INTO vpn_domains VALUES(19,2,'gemini.google.com','gemini.google.com');
INSERT INTO vpn_domains VALUES(20,2,'generativelanguage.googleapis.com','API principal');
INSERT INTO vpn_domains VALUES(21,2,'aistudio.google.com','aistudio.google.com');
INSERT INTO vpn_domains VALUES(22,2,'alkali.google.com','alkali.google.com');
INSERT INTO vpn_domains VALUES(23,2,'omarchy.org','omarchy.org');
INSERT INTO vpn_domains VALUES(24,2,'signup.cloud.oracle.com','signup.cloud.oracle.com');
INSERT INTO vpn_domains VALUES(25,2,'cloud.oracle.com','cloud.oracle.com');
INSERT INTO vpn_domains VALUES(26,2,'oracle.com','oracle.com');
INSERT INTO vpn_domains VALUES(27,2,'forum.centos-webpanel.com','forum.centos-webpanel.com');
INSERT INTO vpn_domains VALUES(28,2,'centos-webpanel.com','centos-webpanel.com');
INSERT INTO vpn_domains VALUES(29,2,'platform.minimax.io','platform.minimax.io');
INSERT INTO vpn_domains VALUES(30,2,'minimax.io','minimax.io');
INSERT INTO vpn_domains VALUES(31,2,'www.wminimax.io','www.wminimax.io');
INSERT INTO vpn_domains VALUES(32,2,'chat.qwen.ai','chat.qwen.ai');
INSERT INTO vpn_domains VALUES(33,2,'qwen.ai','qwen.ai');
INSERT INTO vpn_domains VALUES(34,2,'www.qwen.ai','www.qwen.ai');
INSERT INTO vpn_domains VALUES(35,2,'commandcode.ai','commandcode.ai');
PRAGMA writable_schema=ON;
CREATE TABLE IF NOT EXISTS sqlite_sequence(name,seq);
DELETE FROM sqlite_sequence;
INSERT INTO sqlite_sequence VALUES('vpn_connections',2);
INSERT INTO sqlite_sequence VALUES('vpn_routes',66);
INSERT INTO sqlite_sequence VALUES('vpn_domains',35);
PRAGMA writable_schema=OFF;
COMMIT;
