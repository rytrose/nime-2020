const db = firebase.firestore();
const rtdb = firebase.database();
const auth = firebase.auth();
const ui = new firebaseui.auth.AuthUI(auth);

let users = db.collection('users');
let rooms = db.collection('rooms');

let socket;

const uiConfig = {
    'callbacks': {
        'signInSuccessWithAuthResult': (authResult, _) => {
            if (authResult.user) {
                console.log("user signed in:", authResult.user);
            }
            if (authResult.additionalUserInfo) {
                console.log("is new user:", authResult.additionalUserInfo.isNewUser);
            }

            // Do not redirect
            return false;
        },
        'signInFailure': (error) => {
            console.log("unable to sign in:", error);
        },
        'uiShown': () => {
            $("#loading").hide();
        }
    },
    'signInOptions': [
        // Only allow email/password sign-in
        {
            provider: firebase.auth.EmailAuthProvider.PROVIDER_ID
        }
    ],
    'credentialHelper': firebaseui.auth.CredentialHelper.NONE
}
ui.disableAutoSignIn();


// Create ephemeral user
// firebase.auth().signInAnonymously().catch((error) => {
//     console.log("error signing in anonymously:", error);
// });

// Once logged in
firebase.auth().onAuthStateChanged((user) => {
    if (user) {
        console.log("user signed in", user);
        $("#content").show();
        $("#content").append(`<p>Hello ${user.displayName} (${user.email})!</p>`);
        // // Add user to firestore
        // users.add({}).then((userDoc) => {
        //     let userID = userDoc.id;
        //     $("#welcome").html(`Welcome, ${userID}!`);
        //     console.log("added user");

        //     // Set RTDB online key
        //     let presenceRef = rtdb.ref(`/status/${userID}`)
        //     presenceRef.set('online').then(() => {
        //         console.log("added user presence ref");
        //     }).catch((error) => {
        //         console.log("error adding user presence ref:", error);
        //     });
            
        //     // On RTDB set offline for deletion from firestore
        //     $(window).on('beforeunload', (e) => {
        //         presenceRef.set('offline')
        //             .then(() => console.log("set user offline"))
        //             .catch((error) => console.error("unable to set user offline:", error));
        //         e.returnValue = "";
        //         return "";
        //     });

        //     // Connect to server via websocket
        //     socket = new Socket(`${location.host}/ws`);

        //     // Set up callbacks
        //     socket.register("operationUpdate", (m) => {
        //         $("#operations").append(`<p>${JSON.stringify(m.operation)}</p>`);
        //     });
        //     socket.register("clearState", (m) => {
        //         $("#operations").empty();
        //     });
        //     socket.register("numMembersUpdate", (m) => {
        //         $("#numMembers").text(m.numMembers);
        //     });

        //     // Announce this user 
        //     socket.addEventListener("open", () => {
        //         socket.send({
        //             "id": uuidv4(),
        //             "type": "announce",
        //             "userID": userID
        //         });
        //     });
        // }).catch((error) => {
        //     console.log("error adding user:", error);
        // });
    } else {
        console.log("authStateChanged, no user signed in");
        $("#loading").hide();
        ui.start('#firebaseui-auth-container', uiConfig);
    }
});

let signOut = () => {
    auth.signOut();
    $("#content").hide();
};

// Sync rooms
// rooms.where('active', '==', true).onSnapshot((snapshot) => {
//     if (!snapshot.size) console.log("No rooms.");

//     snapshot.docChanges().forEach(function (change) {
//         let roomName = change.doc.id;
//         let roomData = change.doc.data();
//         if (change.type === 'removed') $(`#r${change.doc.id}`).remove();
//         else if(change.type === 'added') {
//             $("#rooms").append(`
//             <div id="r${roomName}">
//                 <h3 id="r${roomName}_name">${roomName}</h3>
//                 <p id="r${roomName}_description">${roomData.description}</p>
//                 <button id="r${roomName}_enter">Enter ${roomName}</button>
//                 <hr>
//             </div>
//             `);
//             $(`#r${roomName}_enter`).click(() => {
//                 socket.sendWithResponse({
//                     "id": uuidv4(),
//                     "type": "enterRoom",
//                     "roomName": roomName
//                 })
//                     .then(res => {
//                         $("#currentRoom").html(`
//                         <h3>Welcome to room ${res.roomDoc.RoomName}!</h3>
//                         <p><span id="numMembers">${res.roomDoc.NumMembers}</span> people in this room</p>
//                         <button id="operate">Commit operation</button>
//                         <button id="exitRoom">Exit ${res.roomDoc.RoomName}</button>
//                         <div id="operations">Most recent operations:</div>
//                         `)
//                         for(let operation of res.operations) {
//                             $("#operations").append(`<p>${JSON.stringify(operation)}</p>`)
//                         }
//                         $("#operate").click(() => {
//                             let operation = {
//                                 "operationType": "foo",
//                                 "data": Math.random()
//                             };
//                             $("#operations").append(`<p>${JSON.stringify(operation)}</p>`)
//                             socket.send({
//                                 "type": "operation",
//                                 "roomName": roomName,
//                                 "operation": operation
//                             });
//                         });
//                         $("#exitRoom").click(() => socket.sendWithResponse({
//                                 "id": uuidv4(),
//                                 "type": "exitRoom",
//                                 "roomName": roomName,
//                             })
//                             .then(() => {
//                                 $("#currentRoom").empty();
//                                 $("button[id$=_enter]").prop("disabled", false);
//                             })
//                             .catch(e => {
//                                 console.error("unable to exit room:", e)
//                             })
//                         )
//                     })
//                     .catch(e => console.log("unable to enter room:", e));
//                 $("button[id$=_enter]").prop("disabled", true);
//             })
//         } else {
//             $(`#r${roomName}_description`).html(roomData.description);
//         }
//     });
// });
