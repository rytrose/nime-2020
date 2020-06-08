const db = firebase.firestore();
const rtdb = firebase.database();
const auth = firebase.auth();

let users = db.collection('users');
let rooms = db.collection('rooms');

let socket;
let userDoc;
let lastOperationsCol;
let lastOperations = {};

// Once DOM is ready
$(() => {
    // Set display name logic
    $("#displayNameForm").submit(e => {
        e.preventDefault();
        let user = auth.currentUser;
        if (!user) return;
        let displayName = $("#displayName").val();
        user.updateProfile({
            displayName: displayName
        })
            .then(() => {
                console.log("successfully set display name");
                $("#user").text(user.displayName);
                $("#setDisplayName").hide();
                $("#content").show();
            })
            .catch(error => console.log(error.message));
    });
});

// Once logged in
firebase.auth().onAuthStateChanged(user => {
    if (user) {
        console.log("user signed in", user);
        if (user.displayName) {
            $("#loading").hide();
            $("#user").text(`${user.displayName}`);
            $("#content").show();
        } else {
            $("#loading").hide();
            $("#setDisplayName").show();
        }

        // Retain reference to user doc in firestore
        userDoc = users.doc(user.uid);
        lastOperationsCol = userDoc.collection("lastOperations");

        // Get lastOperations timestamps
        lastOperationsCol.get()
            .then(querySnapshot => {
                for (let docSnapshot of querySnapshot.docs) {
                    let data = docSnapshot.data();
                    lastOperations[docSnapshot.id] = data.lastOperation;
                }
            })
            .catch(e => {
                console.log(`unable to get user doc info: ${e.message}`);
            });

        if (!socket) {
            // Connect to server via websocket
            socket = new Socket(`${location.host}/ws`);

            // Set up callbacks
            socket.register("operationUpdate", m => {
                $("#operations").append(`<p>${JSON.stringify(m.operation)}</p>`);
            });
            socket.register("clearState", m => {
                $("#operations").empty();
            });
            socket.register("numMembersUpdate", m => {
                $("#numMembers").text(m.numMembers);
            });

            // Announce this user
            socket.addEventListener("open", () => {
                socket.sendWithResponse({
                    "id": uuidv4(),
                    "type": "announce",
                    "userID": user.uid
                })
                    .then(res => {
                        if(res.error) {
                            // Client already connected
                            alert("You're already logged in in another tab.");
                            $("#content").remove();
                        }
                    })
                    .catch(error => console.log("unable to announce:", error));
            });
        }
    } else {
        console.log("no user signed in, creating user");
        firebase.auth().signInAnonymously().catch(error => console.log("error signing in anonymously:", error));
        $("#loading").hide();
        $("#setDisplayName").show();
    }
});

// Sync rooms
rooms.where('active', '==', true).onSnapshot(snapshot => {
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
                        <p><span id="numMembers">${res.roomDoc.NumMembers}</span> people in this room</p>
                        <button id="operate">Commit operation</button>
                        <button id="exitRoom">Exit ${res.roomDoc.RoomName}</button>
                        <div id="operations">Most recent operations:</div>
                        `)
                        for(let operation of res.operations) {
                            $("#operations").append(`<p>${JSON.stringify(operation)}</p>`)
                        }
                        $("#operate").click(() => {
                            if (lastOperations[roomName]) {
                                if (Date.now() < lastOperations[roomName] + roomData.submitTimeout) {
                                    alert("cannot operate again yet!");
                                    return;
                                }
                            }

                            let operation = {
                                "operationType": "foo",
                                "data": Math.random()
                            };
                            $("#operations").append(`<p>${JSON.stringify(operation)}</p>`);
                            socket.send({
                                "type": "operation",
                                "roomName": roomName,
                                "operation": operation
                            });

                            // Update last operation timestamp
                            lastOperations[roomName] = Date.now();
                            lastOperationsCol.doc(roomName).set({
                                lastOperation: lastOperations[roomName]
                            })
                                .catch(e => {
                                    console.log(`unable to update lastOperation timestamp: ${e.message}`);
                                });
                        });
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
                    .catch(error => console.log("unable to enter room:", error));
                $("button[id$=_enter]").prop("disabled", true);
            })
        } else {
            $(`#r${roomName}_description`).html(roomData.description);
        }
    });
});
