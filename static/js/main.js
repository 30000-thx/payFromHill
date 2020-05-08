class Steps{
    constructor(wizard){
      this.wizard = wizard;
      this.steps = this.getSteps();
      this.stepsQuantity = this.getStepsQuantity();
      this.currentStep = 0;
    }
    
    setCurrentStep(currentStep){
      this.currentStep = currentStep;
    }
    
    getSteps(){
      return this.wizard.getElementsByClassName('step');
    }
    
    getStepsQuantity(){
      return this.getSteps().length;
    }
    
    handleConcludeStep(){
      this.steps[this.currentStep].classList.add('-completed');
    }
    
    handleStepsClasses(movement){
      if(movement > 0)
        this.steps[this.currentStep - 1].classList.add('-completed');
      else if(movement < 0)
        this.steps[this.currentStep].classList.remove('-completed');  
    }
  }
  
  class Panels{
    constructor(wizard){
      this.wizard = wizard;
      this.panelWidth = this.wizard.offsetWidth;
      this.panelsContainer = this.getPanelsContainer();
      this.panels = this.getPanels();
      this.currentStep = 0;
      
      this.updatePanelsPosition(this.currentStep);
      this.updatePanelsContainerHeight();
    }
    
    getCurrentPanelHeight(){
      return `${this.getPanels()[this.currentStep].offsetHeight}px`;
    }
    
    getPanelsContainer(){
      return this.wizard.querySelector('.panels');
    }
    
    getPanels(){
      return this.wizard.getElementsByClassName('panel');
    }
    
    updatePanelsContainerHeight(){
      this.panelsContainer.style.height = this.getCurrentPanelHeight();
    }
    
    updatePanelsPosition(currentStep){
      const panels = this.panels;
      const panelWidth = this.panelWidth;
      
      for (let i = 0; i < panels.length; i++) {
        panels[i].classList.remove(
           'movingIn',
           'movingOutBackward',
           'movingOutFoward'
        );
          
        if(i !== currentStep){
          if(i < currentStep) panels[i].classList.add('movingOutBackward');
          else if( i > currentStep ) panels[i].classList.add('movingOutFoward');
        }else{
          panels[i].classList.add('movingIn');
        }
      }
      
      this.updatePanelsContainerHeight();
    }
    
    setCurrentStep(currentStep){
      this.currentStep = currentStep;
      this.updatePanelsPosition(currentStep);
    }
  }
  
  class Wizard{
    constructor(obj){
      this.wizard = obj;
      this.panels = new Panels(this.wizard);
      this.steps = new Steps(this.wizard);
      this.stepsQuantity = this.steps.getStepsQuantity();
      this.currentStep = this.steps.currentStep;
      
      this.concludeControlMoveStepMethod = this.steps.handleConcludeStep.bind(this.steps);
      this.wizardConclusionMethod = this.handleWizardConclusion.bind(this);
    }
    
    updateButtonsStatus(){
      if(this.currentStep === 0)
        this.previousControl.classList.add('disabled');
      else
        this.previousControl.classList.remove('disabled');
    }
    
    updtadeCurrentStep(movement){   
      this.currentStep += movement;
      this.steps.setCurrentStep(this.currentStep);
      this.panels.setCurrentStep(this.currentStep);
      
      this.handleNextStepButton();
      this.updateButtonsStatus();
    }
    
    setCurrentStep(step){   
        if ( step >= this.stepsQuantity ) {
            throw('This was an invalid movement');
        }
        this.currentStep = step;
        this.steps.setCurrentStep(this.currentStep);
        this.panels.setCurrentStep(this.currentStep);
        
        this.handleNextStepButton();
        this.updateButtonsStatus();
    }
    
    handleNextStepButton(){   
      if(this.currentStep === this.stepsQuantity - 1){      
        this.nextControl.innerHTML = 'Conclude!';
        
        this.nextControl.removeEventListener('click', this.nextControlMoveStepMethod);
        this.nextControl.addEventListener('click', this.concludeControlMoveStepMethod);
        this.nextControl.addEventListener('click', this.wizardConclusionMethod);
      }else{
        this.nextControl.innerHTML = 'Next';
        
        this.nextControl.addEventListener('click', this.nextControlMoveStepMethod);
        this.nextControl.removeEventListener('click', this.concludeControlMoveStepMethod);
        this.nextControl.removeEventListener('click', this.wizardConclusionMethod);
      }
    }
    
    handleWizardConclusion(){
        document.querySelector(".wizard__congrats-message").style.display = "block";
      this.wizard.classList.add('completed');
    };
    
    addControls(previousControl, nextControl){
      this.previousControl = previousControl;
      this.nextControl = nextControl;
      this.previousControlMoveStepMethod = this.moveStep.bind(this, -1);
      this.nextControlMoveStepMethod = this.moveStep.bind(this, 1);
      
      previousControl.addEventListener('click', this.previousControlMoveStepMethod);
      nextControl.addEventListener('click', this.nextControlMoveStepMethod);
      
      this.updateButtonsStatus();
    }
    
    moveStep(movement){
      if(this.validateMovement(movement)){
        this.updtadeCurrentStep(movement);
        this.steps.handleStepsClasses(movement);
      }else{
         throw('This was an invalid movement');
      }
    }
    
    validateMovement(movement){
      const fowardMov = movement > 0 && this.currentStep < this.stepsQuantity - 1;
      const backMov = movement < 0 && this.currentStep > 0;
      
      return fowardMov || backMov;
    }
  }
  
  let wizardElement = document.getElementById('wizard');
  let wizard = new Wizard(wizardElement);
  let buttonNext = document.querySelector('.next');
  let buttonPrevious = document.querySelector('.previous');
  
  wizard.addControls(buttonPrevious, buttonNext);

function checkConfig() {
    if ( document.querySelector("input[name='accountNumber']").value != "" && document.querySelector("input[name='total']").value != "" ) {
        document.querySelector("#config_next").style.display = "block";
    } else {
        document.querySelector("#config_next").style.display = "none";
    }
}

var submitConfig = function() {
            
    var data = new FormData();
    data.append('accountNumber', document.querySelector("input[name='accountNumber']").value);
    data.append('total', document.querySelector("input[name='total']").value);
    
    const xhr = new XMLHttpRequest();
    const url='/start';

    xhr.open("POST", url);
    xhr.onreadystatechange = (e) => {
        if (xhr.readyState === XMLHttpRequest.DONE && xhr.status == 200) {
            var jsonResponse = JSON.parse(xhr.responseText);
            console.log(jsonResponse);
            if ( jsonResponse.message == 'ok' ) {
                wizard.nextControlMoveStepMethod();
                checkLogin();
            } 
        }
    }

    console.log("submiting config");
    xhr.send(data);

}

document.querySelector("#config_next").addEventListener('click', submitConfig);

function checkLogin() {
    const xhr = new XMLHttpRequest();
    const url='/getStatus';

    xhr.open("GET", url);
    xhr.onreadystatechange = (e) => {
        if (xhr.readyState === XMLHttpRequest.DONE && xhr.status == 200) {
            var jsonResponse = JSON.parse(xhr.responseText);
            console.log(jsonResponse);
            if ( jsonResponse.isLogon ) {
                document.querySelector('.circle-loader').classList.add("load-complete");
                document.querySelector('.login-checkmark').style.display = 'block';
                
                setTimeout(() => {  
                    wizard.nextControlMoveStepMethod(); 
                    document.querySelector('#startPaymentBtn').disabled = false;
                }, 2000);                
            } else {
                setTimeout(() => {  checkLogin(); }, 2000);  
                
            }
        }
    }

    xhr.send();
}

function startPayment() {
    const xhr = new XMLHttpRequest();
    const url='/startPayment';

    xhr.open("GET", url);
    xhr.onreadystatechange = (e) => {
        if (xhr.readyState === XMLHttpRequest.DONE && xhr.status == 200) {
            var jsonResponse = JSON.parse(xhr.responseText);
            console.log(jsonResponse);
            if ( jsonResponse.result ) {
                
                document.querySelector("#startPaymentBtn").style.display = 'none';
                
                //join ws for getting result
                var paymentListener = new WebSocket("ws://"+document.location.host+"/ws", );
                paymentListener.onopen = function (event) {
                    paymentListener.send("Here's some text that the server is urgently awaiting!"); 
                };

                document.querySelector('#payMessage').style.display = 'block';
                paymentListener.onmessage = function (event) {

                    var data = JSON.parse(event.data);

                    if ( data.LastPayRecord != undefined ) {
                        var amountToast = document.createElement("p");
                        amountToast.className = "toast-text";
                        amountToast.innerText = "+ " + (data.LastPayRecord.PayAmount / 100.0)

                        document.querySelector('#payingAmount').appendChild(amountToast);

                        setTimeout(() => {  
                            amountToast.remove()
                            document.querySelector("#paidAmount").innerText = data.PaidAmount / 100.0
                         }, 5000);  
                    } else {

                        document.querySelector("#total").innerText = data.Total / 100.0
                        document.querySelector("#paidAmount").innerText = data.PaidAmount / 100.0
                    }

                    if ( data.IsEnd ) {
                        paymentListener.close();
                        wizard.concludeControlMoveStepMethod();
                        wizard.wizardConclusionMethod();
                    }
                    // console.log(event.data);
                }

            } else {
                // some error meesage
            }
        }
    }

    xhr.send();
}

