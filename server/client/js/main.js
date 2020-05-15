const db = firebase.firestore();
const rtdb = firebase.database();
const auth = firebase.auth();

let users = db.collection('users');
let rooms = db.collection('rooms');

let socket;
let userDoc;
let operationWaitUntil;

// Once DOM is ready
$(() => {
    // Sign in logic
    $("#signInForm").submit(e => {
        e.preventDefault();
        let email = $("#signInEmail").val();
        let password = $("#signInPassword").val();
        auth.signInWithEmailAndPassword(email, password)
            .catch((error) => {
                switch (error.code) {
                    case "auth/invalid-email":
                        alert("Invalid email.");
                        break;
                    case "auth/user-disabled":
                        alert("User disabled.");
                        break;
                    case "auth/user-not-found":
                        $("#signUpEmail").val(email);
                        $("#signUpPassword").val(password);
                        $("#signIn").hide();
                        $("#signUp").show();
                        break;
                    case "auth/wrong-password":
                        alert("Incorrect password.");
                        break;
                    default:
                        alert(error.message);
                }
            });
    });
    $("#signInLink").click(() => {
        $("#signUp").hide();
        $("#signIn").show();

    });

    // Sign up logic
    $("#signUpForm").submit(e => {
        e.preventDefault();
        let email = $("#signUpEmail").val();
        let password = $("#signUpPassword").val();
        let displayName = $("#signUpDisplayName").val();

        auth.createUserWithEmailAndPassword(email, password)
            .then(userCredential => {
                userCredential.user.updateProfile({
                    displayName: displayName
                })
                    .then(() => { 
                        console.log("successfully set display name");
                        $("#user").text(`${userCredential.user.displayName} (${userCredential.user.email})`);
                        $("#content").show();
                    })
                    .catch(e => { console.log(error.message) });
            })
            .catch(error => {
                switch (error.code) {
                    case "auth/email-already-in-use":
                        alert("Email already in use.");
                        break;
                    case "auth/invalid-email":
                        alert("Invalid email.");
                        break;
                    case "auth/weak-password":
                        alert("Password too weak.");
                        break;
                    default:
                        alert(error.message);
                }
            });
    });
    $("#signUpLink").click(e => {
        e.preventDefault();

        $("#signIn").hide();
        $("#signUp").show();
    });

    // Sign out logic
    $("#signOut").click(e => {
        e.preventDefault();

        auth.signOut();
        $("#content").hide();
        $("#signIn").show();
    });
});

// Once logged in
firebase.auth().onAuthStateChanged(user => {
    if (user) {
        console.log("user signed in", user);
        $("#signIn").hide();
        $("#signUp").hide();
        if (user.displayName) {
            $("#user").text(`${user.displayName} (${user.email})`);
            $("#content").show();
        }

        // Retain reference to user doc in firestore
        userDoc = users.doc(user.uid);

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
    } else {
        console.log("authStateChanged, no user signed in");
        $("#signIn").show();
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
                        <p><span id="numMembers">${res.roomDoc.NumMembers}</span> people in this room</p>
                        <button id="operate">Commit operation</button>
                        <button id="exitRoom">Exit ${res.roomDoc.RoomName}</button>
                        <div id="operations">Most recent operations:</div>
                        `)
                        for(let operation of res.operations) {
                            $("#operations").append(`<p>${JSON.stringify(operation)}</p>`)
                        }
                        $("#operate").click(() => {
                            let operation = {
                                "operationType": "foo",
                                "data": Math.random()
                            };
                            $("#operations").append(`<p>${JSON.stringify(operation)}</p>`)
                            socket.send({
                                "type": "operation",
                                "roomName": roomName,
                                "operation": operation
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
