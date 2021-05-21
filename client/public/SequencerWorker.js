const PATTERN_UPDATE = 'PATTERN_UPDATE';
const PATTERN_MOVED = 'PATTERN_MOVED';
const PATTERN_DELETED= 'PATTERN_DELETED';
const TRACK_DELETED = 'TRACK_DELETED';
const PARAMETER_UPDATE = 'PARAMETER_UPDATE';
const MIDI_EFFECTS_UPDATE = 'MIDI_EFFECTS_UPDATE';
const MIDI_EFFECTS_CLEARED = 'MIDI_EFFECTS_CLEARED';
const SCHEDULE_TRACK_TRIGGERS = 'SCHEDULE_TRACK_TRIGGERS';
const TRACK_DATA_UPDATE = 'TRACK_DATA_UPDATE';

let tracks = {};
let parameters = {};

this.onmessage = function(e){
    let data = e.data.data;
    if ('trackNumber' in data) {
        if (!(data.trackNumber in tracks)) {
            tracks[data.trackNumber] = {
                patterns: {},
            };
        }
    }
    switch(e.data.command){
    case PATTERN_UPDATE:
        if (tracks[data.trackNumber].patterns[data.patternNumber] === undefined) {
            tracks[data.trackNumber].patterns[data.patternNumber] = {};
        }
        tracks[data.trackNumber].patterns[data.patternNumber].matrix = data.matrix;
        tracks[data.trackNumber].patterns[data.patternNumber].patternLength = data.patternLength;
        break;
    case PARAMETER_UPDATE:
        parameters[data.id] = data;
        break;
    case MIDI_EFFECTS_UPDATE:
        let chain = new MidiEffectsChain(data.trackNumber);
        chain.fromJson(data);
        tracks[data.trackNumber].midiEffects = chain;
        break;
    case TRACK_DATA_UPDATE:
        tracks[data.trackNumber].trackData = data.trackData;
    case MIDI_EFFECTS_CLEARED:
        tracks[data.trackNumber].midiEffects.clear();
    case SCHEDULE_TRACK_TRIGGERS:
        scheduleTriggers(data.step, data.patternNumber);
        break;
    }
};

function scheduleTriggers(step, patternNumber) {
    let list = [];
    for (let trackNumber in tracks) {
        let track = tracks[trackNumber];
        if (track.patterns[patternNumber] === undefined) {
            continue;
        }
        let patternStep = step % track.patterns[patternNumber].patternLength;
        list = list.concat(scheduleTrackTriggers(trackNumber, patternStep, patternNumber));
    }

    let msg = {command: 'TRIGGERS', triggers: list};
    this.postMessage(msg);
}

function liveTrigger(trackNumber, stepData) {
    // send message
}

function liveTriggerRelease(trackNumber, transpose) {
    // send message
}

class StepData {
    constructor() {
        this.on = undefined;
        this.reverse = undefined;
        this.transpose = undefined;
        this.duration = undefined;
        this.velocity = undefined;
        this.chop = undefined;
        this.chopDuration = undefined;
        this.attack = undefined;
        this.release= undefined;
        this.slice = undefined;
        this.panning = undefined;

        this.pitchAttack = undefined;
        this.pitchDecay = undefined;
        this.pitchSustain = undefined;
        this.pitchRelease = undefined;
        this.monophonic = undefined;
        this.selectedRecording = undefined;
    }

    copyFrom(stepData) {
        this.on = stepData.on;
        this.reverse = stepData.reverse;
        this.transpose = stepData.transpose;
        this.duration = stepData.duration;
        this.velocity = stepData.velocity;
        this.chop = stepData.chop;
        this.chopDuration = stepData.chopDuration;
        this.attack = stepData.attack;
        this.release= stepData.releas;
        this.slice = stepData.slice;
        this.panning = stepData.panning;

        this.pitchAttack = stepData.pitchAttack;
        this.pitchDecay = stepData.pitchDecay;
        this.pitchSustain = stepData.pitchSustain;
        this.pitchRelease = stepData.pitchRelease;
        this.monophonic = stepData.monophonic;
        this.selectedRecording = stepData.selectedRecording;
        this.polyphonicSteps = stepData.polyphonicSteps.map(
            step => ({... step}));
    }
}

function getCopy(toCopy) {
    let stepData = new StepData();
    stepData.on = toCopy.on;
    stepData.reverse = toCopy.reverse;
    stepData.transpose = toCopy.transpose;
    stepData.duration = toCopy.duration;
    stepData.velocity = toCopy.velocity;
    stepData.chop = toCopy.chop;
    stepData.chopDuration = toCopy.chopDuration;
    stepData.attack = toCopy.attack;
    stepData.release= toCopy.releas;
    stepData.slice = toCopy.slice;
    stepData.panning = toCopy.panning;
    stepData.pitchAttack = toCopy.pitchAttack;
    stepData.pitchDecay = toCopy.pitchDecay;
    stepData.pitchSustain = toCopy.pitchSustain;
    stepData.pitchRelease = toCopy.pitchRelease;
    stepData.monophonic = toCopy.monophonic;
    stepData.selectedRecording = toCopy.selectedRecording;
    stepData.polyphonicSteps = toCopy.polyphonicSteps.map(
        x => getCopy(x));
    return stepData;
}
function getStepDataForStep(trackNumber, originalStepData) {
    let trackData = tracks[trackNumber].trackData;
    let copy = getCopy(originalStepData);
    let polyphony = [copy].concat(copy.polyphonicSteps);

    for (let i in polyphony) {
        let stepData = polyphony[i];
        stepData.duration = stepData.duration === null ? trackData.duration : stepData.duration;
        stepData.transpose = stepData.transpose === null ? trackData.transpose : stepData.transpose;
        stepData.velocity = stepData.velocity == null ? trackData.velocity : stepData.velocity;
        stepData.chop = stepData.chop == null ? trackData.chop : stepData.chop;
        stepData.chopDuration = stepData.chopDuration == null ? trackData.chopDuration : stepData.chopDuration;
        stepData.attack = stepData.attack == null ? trackData.attack : stepData.attack;
        stepData.release = stepData.release == null ? trackData.release : stepData.release;
        stepData.slice = stepData.slice == null ? trackData.slice : stepData.slice;
        stepData.panning = stepData.panning ;
        stepData.pitchAttack = stepData.pitchAttack === null ? trackData.pitchAttack: stepData.pitchAttack;
        stepData.pitchDecay = stepData.pitchDecay == null ? trackData.pitchDecay: stepData.pitchDecay;
        stepData.pitchSustain = stepData.pitchSustain == null ? trackData.pitchSustain: stepData.pitchSustain;
        stepData.pitchRelease= stepData.pitchRelease == null ? trackData.pitchRelease: stepData.pitchRelease;
        stepData.monophonic = stepData.monophonic == null ? trackData.monophonic : stepData.monophonic;
        stepData.reverse = stepData.reverse == null ? trackData.reverse : stepData.reverse;
        stepData.effectAutomations = stepData.effectAutomations;
    }

    return copy;
}

function scheduleTrackTriggers(trackNumber, step, patternNumber) {
    let track = tracks[trackNumber];
    let pattern = track.patterns[patternNumber].matrix;

    track.midiEffects.onNewStep();

    if (pattern[step] === undefined) {
        track.midiEffects.output(step);
        return [];
    }

    let triggers = [];
    let bucket = pattern[step];
    for (let timeInBucket in bucket) {
        let stepData = getStepDataForStep(trackNumber, bucket[timeInBucket]);
        triggers.push({command: 'TRIGGER_AUTOMATIONS', time: timeInBucket, step: step});

        let polyphony = [stepData].concat(stepData.polyphonicSteps);
        for (let i in polyphony) {
            let polyphonicStep = polyphony[i];
            if (!polyphonicStep.on) {
                    // we only trigger steps that are "on"
                    continue;
                }
                
            // send it to the midi effects 
            track.midiEffects.input(
                polyphonicStep, timeInBucket, step);
        }
    }

    let midiEffectTriggers = track.midiEffects.output(step);
    midiEffectTriggers.forEach(
        trigger => triggers.push({command: 'TRIGGER_TRACK', track: trackNumber, step: step,
                                  trigger: trigger}));

    return triggers;
}


function scheduleNothing() {
}

// Represents Resolution of a beat, for example 1/32 notes or 1/32t (triplet notes)
class Resolution {
    constructor(resolution, isTriplet=false) {
        this.resolution = resolution;
        this.isTriplet = isTriplet;
        this.baseResolution = isTriplet ? 48 : 32;
    }

    // if this is a triplet resolution, then returns the non triplet version
    getTripletOrNonTriplet() {
        if (this.isTriplet) {
            return new Resolution(this.resolution, false);
        } else {
            return new Resolution(this.resolution, true);
        }
    }

    equals(o) {
        return this.resolution === o.resolution &&
            this.isTriplet === o.isTriplet;
    }

    numSteps() {
        if (this.isTriplet) {
            return (this.resolution / 32) * 48;
        } else {
            return this.resolution;
        }
    }

    transformStepAtResolution2(step, resolution) {
        return step * this.numSteps() / resolution.numSteps();
    }

    fromJson(json) {
        this.resolution = json.resolution;
        this.baseResolution = json.baseResolution;
        this.isTriplet = json.isTriplet;
        return this;
    }

    getCopy() {
        return new Resolution(this.resolution, this.isTriplet);
    }
}

class MidiEffectsChain {

    constructor(track) {
        this.track = track;
        this.effects = [];

        this.inputedTriggers = [];

        this.noteToTriggers = {};
    }

    addEffect(effect) {
        this.effects.push(effect);
    }

    // input all the triggers that should happen for this 1/32 step,
    // pre midi effects chain
    input(stepData, time, currentStep) {
        this.inputedTriggers.push(
            new MidiTrigger(stepData, time, currentStep));
    }

    // outputs all the triggers that should be played post chain
    output(stepNumber) {
        // calculates the triggers by going through all the effects
        let input = this.inputedTriggers;
        for (let i in this.effects) {
            let effect = this.effects[i];
            if (!effect.isEnabled()) {
                continue;
            }
            for (let j in input) {
                let trigger = input[j];
                effect.input(trigger);
            }
            input = effect.output(stepNumber);
        }

        return input;
    }

    liveInputTrigger(stepData) {
        if (this.isRealTime()) {
            // need to keep track of what this stepData
            // becomes after going through all the midi effects,
            // so that we can later release them
            let time = undefined;
            let stepNumber = undefined; 

            let triggers = [
                new MidiTrigger(stepData, time, stepNumber)
            ];
                   
            for (let i in this.effects) {
                if (!this.effects[i].isEnabled()) {
                    continue;
                }
                for (let j in triggers) {
                    this.effects[i].input(triggers[j]);
                }
                triggers = this.effects[i].output(time, stepNumber);
            }

            triggers.forEach(
                trigger =>
                    liveTrigger(this.track, trigger.stepData)
            );
                             
            this.noteToTriggers[stepData.transpose] = triggers;
        } else {
            this.addNote(stepData.transpose);
        }
    }

    liveInputRelease(transpose) {
        if (this.isRealTime()) {
            let triggers = this.noteToTriggers[transpose];
            for (let i in triggers) {
                liveTriggerRelease(this.track, triggers[i].stepData.transpose);
            }
            delete this.noteToTriggers[transpose];
        } else {
            this.removeNote(transpose);
        }
    }

    // Used when user does a "live" trigger, i.e. uses a midi 
    // keyboad/computer keyboard to trigger a sampler, synth, etc
    addNote(transpose) {
        this.effects[0].addNote(transpose);
    }

    removeNote(transpose) {
        this.effects.forEach(effect => effect.removeNote(transpose));
    }

    isRealTime() {
        return this.effects
            .filter(effect => effect.isEnabled())
            .every(e => e.isRealTime());
    }

    // returns whether whether sequencer should wait
    // until last possible moment to sequence, in case
    // user triggers something at the last second
    isLiveRealTime() {
        /*
        if (!this.track.isEnabled) {
            return false;
        }
        */
        return this.effects.some(e => e.name === 'Arpeggiator');
    }

    // On each new step scheduling in the PlaySequencerEngine, we clear
    // the inputedTriggers
    onNewStep() {
        this.inputedTriggers = [];
    }

    // When "starting" playback of a sequence, we want to
    // clear everything
    clear() {
        this.effects.forEach(
            effect =>
                effect.clear());
    }

    fromJson(json) {
        for (let i in json) {
            let effect = getMidiEffect(json[i], this.track);
            if (effect === undefined) {
                continue;
            }
            effect.fromJson(json[i]);
            this.addEffect(effect);
        }
    }
}

function getMidiEffect(json, track) {
    if (json.name === 'Arpeggiator') {
        let arp = new ArpEffect(track);
        return arp;
    }

    return undefined;
}

const UP = 'UP';
const DOWN = 'DOWN'

const CACHE_SIZE = 16;

class MidiTrigger {
    constructor(stepData, time, stepNumber) {
        this.stepData = stepData;
        this.time = time;
        this.stepNumber = stepNumber;
    }
}
class AbstractEffect {
    constructor(name, track) {
        this.name = name;
        this.track = track;
        this.enabled = false;
    }
    
    toggleEnabled() {
        this.enabled = !this.enabled;
    }

    isEnabled() {
        return this.enabled;
    }

    fromJson(json) {
        this.enabled = json.enabled;
    }
}


class MidiEffect extends AbstractEffect {
    constructor(name, track) {
        super(name, track);
        this.track = track;
    }

    isRealTime() {
        return true;
    }
    
    // must be overridden by sub-classes
    input() {}

    // must be overridden by sub-classes
    output() {}
}

const BASE_RESOLUTION = new Resolution(32, false);

// Util for determining trigger times based on
// resolution and the current step

function getTriggerTimes(stepNumber, resolution) {
    let minStepNumber = resolution.transformStepAtResolution2(stepNumber, BASE_RESOLUTION);
    let maxStepNumber = resolution.transformStepAtResolution2(stepNumber + 1, BASE_RESOLUTION);
    
    let stepsAtResolution = [];
    
    if (Math.floor(minStepNumber) == minStepNumber) {
        stepsAtResolution.push(minStepNumber);
    } else if (maxStepNumber - minStepNumber > 1) {
        stepsAtResolution.push(Math.ceil(minStepNumber));
    }
    
    // is there a whole number between the two numbers, not including the second number?
    if (Math.floor(maxStepNumber) > Math.floor(minStepNumber) &&
        Math.floor(maxStepNumber) != maxStepNumber) {
        
        stepsAtResolution.push(Math.floor(maxStepNumber));
    }
    
    let ret = [];
    // now we want to convert them back to the base resolution
    for (let i=0 ; i < stepsAtResolution.length; i++) {
        let step = stepsAtResolution[i];
        let baseResolutionStep = BASE_RESOLUTION.transformStepAtResolution2(step, resolution);
        let offset = baseResolutionStep - stepNumber;
        ret.push({step: step, time: offset});
    }
    return ret;
}

class ArpEffect extends MidiEffect {
   constructor(track) {
        super('Arpeggiator', track);

        this.initCache();

        this.noteToDuration = {};
        this.notes = [];
        this.index = 0;
        this.repeatIndex = 0;
        this.currentOctave = 0;
   }

    isRealTime() {
        return false;
    }

    clear() {
        this.noteToDuration = {};
        this.notes = [];
        this.repeatIndex = 0;
        this.index = 0;
        if (this.arpDirection === DOWN) {
            this.currentOctave = this.arpOctaves;
        } else {
            this.currentOctave = 0;
        }
    }

    addNote(transpose) {
        let trackData = tracks[this.track].trackData;
        let stepData = this.getStep(trackData, transpose, null, false);
        stepData.velocity = trackData.velocity;
        stepData.duration = null;
        stepData.transpose = transpose;

        this.input(new MidiTrigger(stepData, 0, 0), true);
    }

    // overriden method
    input(trigger, isLive=false) {
        let duration = trigger.stepData.duration;
        let transpose = trigger.stepData.transpose;

        if (duration !== null) {
            if (transpose in this.noteToDuration) {
                this.noteToDuration[transpose] = this.noteToDuration[transpose] + duration*2
            } else {
                this.noteToDuration[transpose] = duration * 2;
            }
        }

        if (this.getIndexOfTrigger(trigger) < 0) {
            this.notes.push(trigger);
            this.notes = this.notes.sort(
                (a,b) => a.stepData.transpose - b.stepData.transpose);

            if (this.arpDirection === DOWN) {
                this.notes.reverse();
            }
        }
    }

    // overriden method
    output(stepNumber) {
        return this.getNextTriggers(stepNumber);
    }

    getIndexOfTrigger(trigger) {
        for (let i in this.notes) {
            if (this.notes[i].stepData.transpose === trigger.stepData.transpose) {
                return i;
            }
        }

        return -1;
    }

    removeNote(transpose) {
        let index = -1;
        for (let i in this.notes) {
            if (transpose === this.notes[i].stepData.transpose) {
                index = i;
                break;
            }
        }
        if (index > -1) {
            this.notes.splice(index, 1);
        }

        if (this.index >= this.notes.length) {
            this.index = 0;
            this.currentOctave = 0;
        }

        if (this.notes.length === 0) {
            this.clear();
        }
    }

    nextNote(octaves) {
        if (this.notes.length == 0) {
            return undefined;
        }

        let trigger = this.notes[this.index];
        let note = trigger.stepData.transpose;
        note = note + 12 * this.currentOctave;
        this.repeatIndex++;
        
        if (this.repeatIndex >= this.arpRepeat) {
            this.repeatIndex = 0;
            this.index++;
        }
        if (this.index >= this.notes.length) {
            this.index = 0;
            if (this.arpDirection === UP) {
                this.currentOctave = (this.currentOctave + 1) % (octaves + 1);
            } else if (this.arpDirection === DOWN) {
                this.currentOctave = (this.currentOctave - 1) % (octaves + 1);
                if (this.currentOctave < 0) {
                    this.currentOctave = octaves;
                }
            }
        } 

        let stepData = this.getStep(
            trigger.stepData,
            note, 
            tracks[this.track].trackData.duration * this.arpDurationFactor);
        return new MidiTrigger(stepData, trigger.time, trigger.stepNumber);
    }

    getNextTriggers(stepNumber) {
        let triggerTimes = getTriggerTimes(stepNumber, this.arpRate);
        let triggers = triggerTimes.map(
            triggerTime => {
                let trigger = this.nextNote(this.arpOctaves);
                if (trigger !== undefined) {
                    trigger.time = triggerTime.time;
                }
                return trigger;
            })
            .filter(
                trigger =>
                    trigger !== undefined);

        // now reduce the duration of each note by 1 and delete
        // if needed
        for (let note in this.noteToDuration) {
            this.noteToDuration[note] -= 1;
            if (this.noteToDuration[note] <= 0) {
                delete this.noteToDuration[note];
                this.removeNote(parseInt(note));
            }
        }

        if (Object.keys(this.noteToDuration).length === 0 &&
           this.notes.length === 0) {
            this.clear();
        }

        return triggers;
    }

    fromJson(json) {
        super.fromJson(json);
        this.arpRate = new Resolution(16, false)
            .fromJson(json.arpRate);
        this.arpOctaves = json.arpOctaves;
        this.arpDirection = json.arpDirection;
        this.arpDurationFactor = json.arpDurationFactor;
    }

    initCache() {
        this.cachedSteps = [];

        for (let i=0; i < CACHE_SIZE; i++) {
            let stepData = new StepData();
            this.cachedSteps.push(stepData);
        }

        this.cacheIndex = 0;
    }

    getStep(step, transpose, duration, useCache=false) {
        let stepToUse = this.cachedSteps[this.cacheIndex];
        stepToUse.copyFrom(step);
        stepToUse.duration = duration;
        stepToUse.transpose = transpose;
        
        this.cacheIndex = (this.cacheIndex + 1) % this.cachedSteps.length;
        return stepToUse;
    }
}
