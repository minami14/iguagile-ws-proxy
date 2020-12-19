var Iguagile = new class {
    constructor() {
    };

    RoomApiClient = class {
        constructor(url) {
            this.baseUrl = url;
        }

        create(request, callback) {
            const xhr = new XMLHttpRequest();
            const url = `${this.baseUrl}/rooms`;
            xhr.open('POST', url, true);
            xhr.setRequestHeader("Content-Type", "application/json");
            xhr.onreadystatechange = function () {
                if (this.readyState === XMLHttpRequest.DONE && this.status === 201) {
                    const resp = JSON.parse(xhr.responseText);
                    const room = resp.result;
                    room.application_name = request.application_name;
                    room.version = request.version;
                    room.password = request.password;
                    callback(room);
                }
            };
            xhr.send(JSON.stringify(request));
        }

        search(request, callback) {
            const xhr = new XMLHttpRequest();
            const url = `${this.baseUrl}/rooms?name=${request.application_name}&version=${request.version}`;
            xhr.open('GET', url, true);
            xhr.onreadystatechange = function () {
                if (this.readyState === XMLHttpRequest.DONE && this.status === 200) {
                    const resp = JSON.parse(xhr.responseText);
                    const rooms = resp.result;
                    rooms.forEach((room, i) => {
                        room.application_name = request.application_name;
                        room.version = request.version;
                    })
                    callback(rooms);
                }
            }
            xhr.send();
        }
    };

    Client = class {
        constructor(url) {
            this.proxyUrl = url
        }

        onConnect = function () {
        };
        onReceive = function (data) {
        };

        connect(room) {
            this.room = room;
            const url = `ws://${this.room.server.server}:${this.room.server.port}`;
            const socket = new WebSocket(this.proxyUrl);
            this.socket = socket;
            const onConnect = this.onConnect;
            const onReceive = this.onReceive;
            socket.addEventListener('open', function (event) {
                socket.send(JSON.stringify(room));
                onConnect();
            });
            this.socket.addEventListener('message', function (event) {
                onReceive(event.data);
            });
        }

        send(data) {
            this.socket.send(data);
        }
    }
}();
