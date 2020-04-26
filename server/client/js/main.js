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
            socket = new Socket("localhost:8000/ws");

            // Set up callbacks
            socket.register("stateUpdate", (m) => {
                $("#testState").html(m.state.test);
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
        let roomID = change.doc.id;
        let roomData = change.doc.data();
        console.log("Room change:", roomID, roomData);
        if (change.type === 'removed') $(`#r${change.doc.id}`).remove();
        else if(change.type === 'added') {
            $("#rooms").append(`
            <div id="r${roomID}">
                <h3 id="r${roomID}_name">${roomID}</h3>
                <p id="r${roomID}_description">${roomData.description}</p>
                <button id="r${roomID}_enter">Enter ${roomID}</button>
                <hr>
            </div>
            `);
            $(`#r${roomID}_enter`).click(() => {
                socket.sendWithResponse({
                    "id": uuidv4(),
                    "type": "enterRoom",
                    "roomID": roomID
                })
                    .then(res => {
                        $("#currentRoom").html(`
                        <h3>Welcome to room ${res.roomData.roomID}!</h3>
                        <p>Test room state: <span id="testState">${res.roomData.test}</span></p>
                        <button id="changeState">Commit operation</button>
                        <button id="exitRoom">Exit ${res.roomData.roomID}</button>
                        `)
                        $("#changeState").click(() => socket.send({
                            "type": "operation",
                            "roomID": roomID,
                        }))
                        $("#exitRoom").click(() => socket.sendWithResponse({
                                "id": uuidv4(),
                                "type": "exitRoom",
                                "roomID": roomID,
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
            $(`#r${roomID}_description`).html(roomData.description);
        }
    });
});
