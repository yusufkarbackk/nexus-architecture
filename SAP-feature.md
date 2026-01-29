## SAP Feature

- saat klik edge yang terkoneksi ke node SAP akan muncul panel config
- panel config akan menampilkan schema dan table yang ada di SAP
- panel config juga menampilkan transformasi data seperti yang ada di panel desitnation database
- bisa select, insert, update, delete ke table SAP

- untuk response dari query SELECT akan berbentuk JSON yang bisa mendukung pagination dan di return untuk digunakan kembali di aplikasi customer. 
- selama proses SELECT nexus akan menunggu hasil dari query tersebut dengan koneksi yang idle tanpa timeout, mungkin menggunakan websocket
- sebelumnya untuk  sender app yang akan mengirim data ke database tujuan atau ke REST API maka akan mengirim ke nexus agent yang di install di server customer, apakah akan berbeda untuk SELECT data dari SAP atau database tujuan?
- 
