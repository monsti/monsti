// Enable list action order dragging
(function() {
  var nodeTable = document.querySelector(".node-table tbody");
  var orderColumns = nodeTable.querySelectorAll(".node-table-order");
  var getOrderInput = function(row) {
    return row.querySelector(".node-table-order input");
  }
  var childRows = [];
  var resetFn = function(changeOrder) {
    childRows = [].slice.call(nodeTable.querySelectorAll("tr"), 0);
    var order = 0;
    for (var i = 0; i < childRows.length; i++) {
      childRows[i].classList.remove("drag-before");
      childRows[i].classList.remove("drag-after");
      childRows[i].classList.remove("drag-over");
      if (changeOrder) {
        getOrderInput(childRows[i]).value = order;
      }
      order += 1;
    }
  }
  resetFn();
  for (var i = 0; i < orderColumns.length; i++) {
//    orderColumns[i].style.display = "none";
  }
  for (var i = 0; i < childRows.length; i++) {
    childRows[i].setAttribute("draggable", "true");

    childRows[i].addEventListener("dragstart", function(event) {
      var current = childRows.indexOf(event.target);
      event.dataTransfer.setData(
        "application/x-monsti-list-order-node", current);
      for (var j = 0; j < childRows.length; j++) {
        if (j < current) {
          childRows[j].classList.add("drag-before");
        }
        if (j > current) {
          childRows[j].classList.add("drag-after");
        }
      }
      nodeTable.classList.add("drag-active");
    }, false);

    childRows[i].addEventListener("dragend", function(event) {
      nodeTable.classList.remove("drag-active");
      resetFn();
    }, false);

    childRows[i].addEventListener("dragenter", function(event) {
      event.target.classList.add("drag-over");
      event.preventDefault();
    }, false);
    childRows[i].addEventListener("dragover", function(event) {
      event.preventDefault();
    }, false);

    childRows[i].addEventListener("dragleave", function(event) {
      event.target.classList.remove("drag-over");
    }, false);

    childRows[i].addEventListener("drop", function(event) {
      var current = childRows.indexOf(event.target);
      var dragged = parseInt(event.dataTransfer.getData(
        "application/x-monsti-list-order-node"));
      if (dragged > current) {
        nodeTable.insertBefore(childRows[dragged], childRows[current]);
      }else if (dragged < current) {
        nodeTable.insertBefore(childRows[dragged], childRows[current].nextSibling);
      }
      resetFn(true);
    }, false);
  }
})();
