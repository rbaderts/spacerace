window.onload = function () {

    function highlight(e) {
        if (selected[0]) selected[0].className = '';
         e.target.parentNode.className = 'selected';
    }

    var table = document.getElementById('races_table'),
        selected = table.getElementsByClassName('selected');
        table.onclick = highlight;

        $("#tst").click(function () {
            var value = $(".selected td:first").html();
            value = value || "No row Selected";
            alert(value);
        });
};


jQuery('#practise_race').on('click', function() {
    jQuery.post( "/newrace", function( data ) {

        console.log("/newrace result = " + JSON.stringify(data))

        var page = "/races/"+data.id
//        jQuery.get(page, function (data) {
            window.location.assign(page)
//        });

        // window.location.href="/races/"+data.id
        //console.log("/race result = " + JSON.stringify(data))
    });
});

function JoinButton(e, id) {
    e.stopPropagation()
    var page = "/races/"+id
 //   jQuery.get(page, function (data) {
         window.location.assign(page)
  //  });
}


function RegisterButton(e, id) {

    e.stopPropagation()

}

function Logout(e, id) {

}
