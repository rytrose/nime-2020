var db = firebase.firestore();
var rtdb = firebase.database();
var auth = firebase.auth();

var users = db.collection('users');
var rooms = db.collection('rooms');


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
            var welcomeElement = $("#welcome").html(`Welcome, ${userDoc.id}!`);
            console.log("added user");

            // Set RTDB online key
            var presenceRef = rtdb.ref(`/status/${userDoc.id}`)
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
                presenceRef.onDisconnect().set('offline');
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
        var roomName = change.doc.id;
        var roomData = change.doc.data();
        console.log("Room change:", roomName, roomData);
        if (change.type === 'removed') $(`#r${change.doc.id}`).remove();
        else if(change.type === 'added') {
            $("#rooms").append(`
            <div id="r${roomName}">
                <h3 id="r${roomName}_name">${roomName}</h3>
                <p id="r${roomName}_description">${roomData.description}</p>
            </div>
            `);
        } else {
            $(`#r${roomName}_description`).html(roomData.description);
        }
    });
});

// Create WebSocket connection.
const socket = new WebSocket('ws://localhost:8000/ws');

// Connection opened
socket.addEventListener('open', function (event) {
    socket.send('Hello Server!');
});

// Listen for messages
socket.addEventListener('message', function (event) {
    console.log('Message from server ', event.data);
});
