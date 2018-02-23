$(document).ready(function () {
    $('#filter').click(function() {
        $.ajax({
            url: $(this).attr('data-url'),
            data: {
               startDate: $('#startDate').val(),
               endDate: $('#endDate').val(),
               keyword: $('#description').val(),
            },
            beforeSend: function() {
                $('.indicator').show();   
            },
            success: function(data) {
                $('#items').html(data);
            },
            error: function (jqXHR, textStatus, errorThrown) {
                console.log("TEXT:", textStatus);        
            },
            complete: function () {
                $('.indicator').hide();
                feather.replace();
            }
        });
    }) ;
});