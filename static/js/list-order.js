// Enable list action order dragging
(function() {
  var nodeList = document.querySelector(".node-list");
  var orderColumns = nodeList.querySelectorAll(".node-list-order");
  var getOrderInput = function(item) {
    return item.querySelector(".node-list-order");
  }
  var childItems = [];
  var resetFn = function(changeOrder) {
    childItems = [].slice.call(nodeList.querySelectorAll("li"), 0);
    var order = 0;
    for (var i = 0; i < childItems.length; i++) {
      childItems[i].classList.remove("drag-before");
      childItems[i].classList.remove("drag-after");
      childItems[i].classList.remove("drag-over");
      childItems[i].classList.remove("drag-start");
      if (changeOrder) {
        getOrderInput(childItems[i]).value = order;
      }
      order += 1;
    }
  }
  resetFn();
  for (var i = 0; i < orderColumns.length; i++) {
//    orderColumns[i].style.display = "none";
  }
  for (var i = 0; i < childItems.length; i++) {
    childItems[i].setAttribute("draggable", "true");

    childItems[i].addEventListener("dragstart", function(event) {
      var current = childItems.indexOf(event.target);
      event.target.classList.add("drag-start");
      event.dataTransfer.setData(
        "application/x-monsti-list-order-node", current);
      for (var j = 0; j < childItems.length; j++) {
        if (j < current) {
          childItems[j].classList.add("drag-before");
        }
        if (j > current) {
          childItems[j].classList.add("drag-after");
        }
      }
      nodeList.classList.add("drag-active");
    }, false);

    childItems[i].addEventListener("dragend", function(event) {
      nodeList.classList.remove("drag-active");
      resetFn();
    }, false);

    childItems[i].addEventListener("dragenter", function(event) {
      event.target.classList.add("drag-over");
      event.preventDefault();
    }, false);
    childItems[i].addEventListener("dragover", function(event) {
      event.preventDefault();
    }, false);

    childItems[i].addEventListener("dragleave", function(event) {
      event.target.classList.remove("drag-over");
    }, false);

    childItems[i].addEventListener("drop", function(event) {
      var current = childItems.indexOf(event.target);
      var dragged = parseInt(event.dataTransfer.getData(
        "application/x-monsti-list-order-node"));
      if (dragged > current) {
        nodeList.insertBefore(childItems[dragged], childItems[current]);
      }else if (dragged < current) {
        nodeList.insertBefore(childItems[dragged], childItems[current].nextSibling);
      }
      resetFn(true);
    }, false);
  }
})();
