/* "processor.js" file. */
class PortProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
      console.log('port process called');
      this.port.onmessage = (event) => {
          // Handling data from the node.
          console.log('port processor');
          console.log(event.data.tracks.length);
      };

      this.port.postMessage('Hi!');
  }

  process(inputs, outputs, parameters) {
    // Do nothing, producing silent output.
    return true;
  }
}

registerProcessor('port-processor', PortProcessor);
