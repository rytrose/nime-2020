var db = firebase.firestore();
var rtdb = firebase.database();
var auth = firebase.auth();

var users = db.collection('users');


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
            console.log("added user");

            // Set RTDB online key
            var presenceRef = rtdb.ref(`/status/${userDoc.id}`)
            presenceRef.set('online').then(() => {
                console.log("added user presence ref");
            }).catch((error) => {
                console.log("error adding user presence ref:", error);
            });
            
            // On RTDB set offline for deletion from firestore
            presenceRef.onDisconnect().set('offline');
        }).catch((error) => {
            console.log("error adding user:", error);
        });
    }
});
