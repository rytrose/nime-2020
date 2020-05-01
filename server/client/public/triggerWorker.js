this.onmessage = function(e){
  switch(e.data.command){
    case 'init':
      console.log('init called');
      console.log(e);
      break;
  default:
      break;
  }
};

