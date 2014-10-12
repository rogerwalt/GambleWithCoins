'use strict';

var WebSocketHandler = {
  isConnected: false,
  webSocket: null,
  connect: function() {
    var ws = new WebSocket("ws://localhost:8080/play/");
    WebSocketHandler.webSocket = ws;
    ws.onopen = function() {
      console.log("WebSocket opened");
      WebSocketHandler.isConnected = true;
      ws.onclose = function() {
        console.log("WebSocket closed");
        WebSocketHandler.ws = null;
        WebSocketHandler.isConnected = false;
      }
    }
  },
  send: function(dataToSend, receiveCallback) {
    // check if connected
    if (WebSocketHandler.isConnected) {
      // send stuff
      WebSocketHandler.webSocket.send(JSON.stringify(dataToSend));
      // set new callback on ws.onmessage -> receiveCallbackFunctionPointer
      WebSocketHandler.webSocket.onmessage = function(message) {
        receiveCallback(JSON.parse(message.data));
      };
    } else {
      console.log("Error: Not yet connected.");
    }
  },
  listen: function(receiveCallback) {
    WebSocketHandler.webSocket.onmessage = function(message) {
      receiveCallback(JSON.parse(message.data));
    };
  }
};

/* Controllers */

function AppCtrl($scope, $q, $rootScope, $timeout) {

var maxCount = 5
WebSocketHandler.connect({});

WebSocketHandler.listen(function(d) {
    if(d.data.command == "signal") {
      signals.push({'player': opponent, 'signal': data.signal});
    }

    if(d.data.command == "endRound") {
      $scope.myAction = null;
      $scope.recentOutcome = data.outcome;
      $scope.recentBalanceDifference = data.balanceDifference;
      $scope.round += 1;
      $scope.initRound();
    }

    if(d.data.command == "endGame") {
      $scope.round = 0;
      $scope.myAction = null;
      $scope.matched = false;
      $scope.round = 0;
    }

    if(d.data.command == "matched") {
      $scope.matched = true;
    }
});

$scope.$watch('endOfRound', function(value) {
  if(value == true) {
    $('.endOfRound').alert();
    console.log("END OF ROUND")
  }
});

$scope.$watch('myAction', function(value) {
  if(value != null && $scope.endOfRound != null) {
    $('.waitForOpponent').alert();
    console.log("WACHTEN OP TEGENSPELER")
  }
});


$(function () {
    $('#info').popover({'html': true});

});

// For the timer
$scope.roundProgressData = {
      label: 0,
      percentage: 0
}

$scope.signalIcons = ['fa-times-circle', 'fa-check-circle', 'fa-smile-o', 'fa-frown-o'];
$scope.signals = [];
$scope.join = false;
$scope.authenticated = false;
$scope.matched = false;
$scope.round = 1;
$scope.balance = 0;
$scope.authenticated = true;
$scope.matched = true;
$scope.join = true;
$scope.endOfRound = false;
$scope.myAction = null;

$scope.initRound = function() {
    $scope.getBalance();
    $scope.endOfRound = false;
    $scope.counter = maxCount;
    $scope.myAction = null;
    $scope.signals = []
}

$scope.getBalance = function() {
  WebSocketHandler.send({'command': 'getBalance'}, function(data) {
    if(data.command == 'balance') {
      console.log(data)
      console.log('balance: ' + data.result);
      $scope.balance = data.result;
    }
  });
};

$scope.login = function(name, password) {
  WebSocketHandler.send({command: 'login', name: name, password: password}, function(data) {
    if(data.result == 'success') {
      $scope.authenticated = true;
      $scope.getBalance();
    }
  });

};

$scope.register = function(name, password) {
  WebSocketHandler.send({command: 'register', name: name, password: password}, function(data) {
    if(data.result == 'success') {
      $scope.authenticated = true;
    }
  });
};


// Player indicates he wants to start a new game
$scope.joinGame = function() {
  WebSocketHandler.send({'command': 'join'});
  $scope.join = true;
}

$scope.performAction = function(action) {
  if($scope.myAction != null) {
    return;
  }
 $scope.myAction = action;
  WebSocketHandler.send({'command': 'action', 'action': action});
}

$scope.sendRequestOnOpen = function(request) {
  return sendRequest(request);
}

$scope.$watch('roundProgressData', function (newValue, oldValue) {
  newValue.percentage = newValue.label / 100;
}, true);

// The counter
$scope.counter = maxCount;

$scope.onTimeout = function(){
    if($scope.counter == 0) {
        $scope.endOfRound = true;
        return;
    }
  
    $scope.counter--;

      if($scope.counter == 0) {
        $scope.endOfRound = true;
      }
      $scope.endOfRound = false;

      mytimeout = $timeout($scope.onTimeout,1000);
}
var mytimeout = $timeout($scope.onTimeout,1000);

$scope.stop = function() {
   $timeout.cancel(mytimeout);
}



/*
// The user gets a confirmation/error on withdraw
socket.on('withdraw', function (data) {
  $scope.withdraw = data.result;
});

// Other player played
socket.on('outcome', function (data) {
  socket.emit('get:balance')
  $scope.outcome = data.result;
});

// Methods published to the scope
// ==============================

// Request a depost address
$scope.getDepositAddress = function() {
  socket.emit('get:deposit:address')
}
}

*/

$scope.sendSignal = function(signal) {
  $scope.signals.push({'player': 'you', 'signal': signal});
  WebSocketHandler.send({command: 'signal', signal: signal}, function(data) {
    console.log(data);
  });
}

$scope.$watchCollection('signals', function() {
  console.log($('div.signaloverview').scrolltop);
  $(".signaloverview").animate({ scrollTop: $('.signaloverview').height()}, 1000);
});

}
