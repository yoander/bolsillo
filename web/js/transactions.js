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
                $('#transTable').html(data);
            },
            complete: function () {
                $('.indicator').hide();
                feather.replace();
            }
        });
    }) ;
});