#! /bin/bash

k8s_extenstions_uninstall(){
  systemctl stop isecl-k8s-controller.service
  systemctl stop isecl-k8s-scheduler.service
  rm /etc/systemd/system/isecl-k8s-controller.service
  rm /etc/systemd/system/isecl-k8s-scheduler.service
  rm -rf /opt/isecl-k8s-extensions
  rm -rf /root/attestation-hub-keystores
  rm /usr/local/bin/isecl-k8s-extensions
  systemctl daemon-reload
}

k8s_extenstions_help(){
  echo "Usage:"
  echo "isecl-k8s-extenstions <action> <component>"
  echo "action: start|stop|status|restart"
  echo "component: custom-controller extended-scheduler"
  
  echo "isecl-k8s-extenstions uninstall"
}

k8s_extenstions_stop(){
  if [ $1 == "custom-controller" ]
  then
    systemctl stop isecl-k8s-controller.service
    return
  fi

  if [ $1 == "extended-scheduler" ]
  then
    systemctl stop isecl-k8s-scheduler.service
    return
  fi

  echo "Usage: isecl-k8s-extensions stop custom-controller"
  echo "       isecl-k8s-extensions stop extended-scheduler"

}

k8s_extenstions_restart(){
  if [ $1 == "custom-controller" ]
  then
    systemctl restart isecl-k8s-controller.service
    return
  fi

  if [ $1 == "extended-scheduler" ]
  then
    systemctl restart isecl-k8s-scheduler.service
    return
  fi

  echo "Usage: isecl-k8s-extensions restart custom-controller"
  echo "       isecl-k8s-extensions restart extended-scheduler"

}

k8s_extenstions_start(){
  if [ $1 == "custom-controller" ]
  then
    systemctl start isecl-k8s-controller.service
    return
  fi

  if [ $1 == "extended-scheduler" ]
  then
    systemctl start isecl-k8s-scheduler.service
    return
  fi

  echo "Usage: isecl-k8s-extensions start custom-controller"
  echo "       isecl-k8s-extensions start extended-scheduler"

}

k8s_extenstions_status(){
  if [ $1 == "custom-controller" ]
  then
    systemctl status isecl-k8s-controller.service
    return
  fi

  if [ $1 == "extended-scheduler" ]
  then
    systemctl status isecl-k8s-scheduler.service
    return
  fi

  echo "Usage: isecl-k8s-extensions status custom-controller"
  echo "       isecl-k8s-extensions status extended-scheduler"
 
}

parse_args() {
  case "$1" in
    help)
      shift
      k8s_extenstions_help
      return $?
      ;;
    uninstall)
      shift
      k8s_extenstions_uninstall
      return $?
      ;;
    start)
      shift
      k8s_extenstions_start $*
      return $?
      ;;
    status)
      shift
      k8s_extenstions_status $*
      return $?
      ;;
    stop)
      shift
      k8s_extenstions_stop $*
      return $?
      ;;
    restart)
      shift
      k8s_extenstions_restart $*
      return $?
      ;;
    *)
      shift
      k8s_extenstions_help      
      return $?
      ;;
  esac
}

parse_args $*
