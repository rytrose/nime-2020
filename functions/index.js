const functions = require('firebase-functions');
const admin = require('firebase-admin');
admin.initializeApp();
const db = admin.firestore();

exports.deleteAnonymousUser = functions.database.ref('/status/{userID}')
    .onUpdate((snapshot, _) => {
        // If RTDB updated to 'offline'
        if (snapshot.after.val() == 'offline') {
            // Delete ephemeral user in firestore
            db.collection('users').doc(snapshot.after.key).delete()
                .then(() => { console.log(`deleted user ${snapshot.after.key} in firestore`) })
                .catch((error) => { console.log(`error deleting user ${snapshot.after.key} in firestore: ${error.message}`) });
            // Delete this reference
            snapshot.after.ref.remove()
                .then(() => { return `successfully cleaned up RTDB entry for ${snapshot.after.key}` })
                .catch((error) => { console.log(`error deleting RTDB entry for ${snapshot.after.key}: ${error.message}`) });
        }
        return "noop - not offline"
    });
