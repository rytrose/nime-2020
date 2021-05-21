// This is "processor.js" file, evaluated in AudioWorkletGlobalScope upon
// audioWorklet.addModule() call in the main global scope.
class RecorderWorklet extends AudioWorkletProcessor {
    constructor() {
        super();
        
        this.recording = false;
        this.port.onmessage = this.handleMessage_.bind(this);
    }
    
    
    clear() {
        this.recBuffers = [[], []];
        this.recLength = 0;
    }

    handleMessage_(event) {
        if (event.data.message === 'record') {
            this.clear();
            this.recording = true;
        } else if (event.data.message === 'stop') {
            this.recording = false;
        } else if (event.data.message === 'get_buffer') {
            let buffer = this.getBuffer();
            // send message
            this.port.postMessage({
                message: {
                    type: 'get_buffer',
                    data: buffer
                }});
                
        }

        console.log('[Processor:Received] ' + event.data.message +
                    ' (' + event.data.contextTimestamp + ')');
    }
    
    process(inputs, outputs, parameters) {
        let input = inputs[0];
        if (this.recording) {
            for (var channel = 0; channel < 2; channel++){
                this.recBuffers[channel].push(input[channel]);
            }
            this.recLength += input[0].length;
        }

        return true;
    }

    getBuffer() {
        var buffers = [];
        for (var channel = 0; channel < 2; channel++){
            buffers.push(this.mergeBuffers(this.recBuffers[channel], this.recLength));
        }

        // then clear to force garbage collection
        this.clear();

        return buffers;
    }

    mergeBuffers(recBuffers, recLength){
        var result = new Float32Array(recLength);
        var offset = 0;
        for (var i = 0; i < recBuffers.length; i++){
            result.set(recBuffers[i], offset);
            offset += recBuffers[i].length;
        }
        return result;
    }


}

registerProcessor('recorder-processor', RecorderWorklet);
