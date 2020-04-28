const db = firebase.firestore();
const rtdb = firebase.database();
const auth = firebase.auth();

let users = db.collection('users');
let rooms = db.collection('rooms');

let socket;


// Create ephemeral user
firebase.auth().signInAnonymously().catch((error) => {
    console.log("error signing in anonymously:", error);
});

// Once logged in
firebase.auth().onAuthStateChanged((user) => {
    if (user) {
        // Add user to firestore
        users.add({
            online: true
        }).then((userDoc) => {
            let userID = userDoc.id;
            $("#welcome").html(`Welcome, ${userID}!`);
            console.log("added user");

            // Set RTDB online key
            let presenceRef = rtdb.ref(`/status/${userID}`)
            presenceRef.set('online').then(() => {
                console.log("added user presence ref");
            }).catch((error) => {
                console.log("error adding user presence ref:", error);
            });
            
            // On RTDB set offline for deletion from firestore
            // On navigate away, or internet connection lost
            presenceRef.onDisconnect().set('offline');
            $(window).on('beforeunload', () => {
                // Handles refresh case
                presenceRef.set('offline');
            });

            // Connect to server via websocket
            socket = new Socket(`${location.host}/ws`);

            // Set up callbacks
            socket.register("operationUpdate", (m) => {
                $("#operations").append(`<p>${JSON.stringify(m.operation)}</p>`);
            });

            // Announce this user 
            socket.addEventListener("open", () => {
                socket.send({
                    "id": uuidv4(),
                    "type": "announce",
                    "userID": userID
                });
            });
        }).catch((error) => {
            console.log("error adding user:", error);
        });
    }
});

// Sync rooms
rooms.where('active', '==', true).onSnapshot((snapshot) => {
    if (!snapshot.size) console.log("No rooms.");

    snapshot.docChanges().forEach(function (change) {
        let roomName = change.doc.id;
        let roomData = change.doc.data();
        if (change.type === 'removed') $(`#r${change.doc.id}`).remove();
        else if(change.type === 'added') {
            $("#rooms").append(`
            <div id="r${roomName}">
                <h3 id="r${roomName}_name">${roomName}</h3>
                <p id="r${roomName}_description">${roomData.description}</p>
                <button id="r${roomName}_enter">Enter ${roomName}</button>
                <hr>
            </div>
            `);
            $(`#r${roomName}_enter`).click(() => {
                socket.sendWithResponse({
                    "id": uuidv4(),
                    "type": "enterRoom",
                    "roomName": roomName
                })
                    .then(res => {
                        $("#currentRoom").html(`
                        <h3>Welcome to room ${res.roomDoc.RoomName}!</h3>
                        <button id="operate">Commit operation</button>
                        <button id="exitRoom">Exit ${res.roomDoc.RoomName}</button>
                        <div id="operations">Most recent operations:</div>
                        `)
                        for(let operation of res.operations) {
                            $("#operations").append(`<p>${JSON.stringify(operation)}</p>`)
                        }
                        $("#operate").click(() => socket.send({
                            "type": "operation",
                            "roomName": roomName,
                            "operation": {
                                "operationType": "foo",
                                "data": Math.random()
                            }
                        }))
                        $("#exitRoom").click(() => socket.sendWithResponse({
                                "id": uuidv4(),
                                "type": "exitRoom",
                                "roomName": roomName,
                            })
                            .then(() => {
                                $("#currentRoom").empty();
                                $("button[id$=_enter]").prop("disabled", false);
                            })
                            .catch(e => {
                                console.error("unable to exit room:", e)
                            })
                        )
                    })
                    .catch(e => console.log("unable to enter room:", e));
                $("button[id$=_enter]").prop("disabled", true);
            })
        } else {
            $(`#r${roomName}_description`).html(roomData.description);
        }
    });
});
