this.onmessage = function(e){
  switch(e.data.command){
    case 'decrypt':
      decrypt(e.data.encryptedContent, e.data.contentKey, e.data.counter, e.data.fromBytes);
      break;
  }
};

let encrypting = 0;

async function decrypt(content, contentKey, counter, fromBytes) {
    const algo = {
        name: "AES-CTR",
        //Don't re-use counters!
        //Always use a new counter every time your encrypt!
        counter: counter,
        length: 64, //can be 1-128
    };

    let dateA = new Date();
    let imported = await crypto.subtle.importKey(
        'raw',
        contentKey,
        {name: 'AES-CTR'},
        false,
        ['encrypt', 'decrypt'],
    );
    let dateB = new Date();
    let time = dateB.getTime() - dateA.getTime();
    console.log('time spent importing = ' + time + 'ms');
 
    encrypting++;
    console.log('encrypting = ' + encrypting + ' many at once bytes=' + content.length);
    let decryptedBytes = await crypto.subtle.decrypt(algo, imported, content);
    let dateC = new Date();
    encrypting--;

    decryptedBytes = new Uint8Array(decryptedBytes);
    let time2 = dateC.getTime() - dateB.getTime();
    console.log("time spent decrypting actually = " + time2);
    this.postMessage(decryptedBytes);
}

