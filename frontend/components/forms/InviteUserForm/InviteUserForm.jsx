import React, { Component, PropTypes } from 'react';

import Button from '../../buttons/Button';
import InputFieldWithIcon from '../fields/InputFieldWithIcon';
import userInterface from '../../../interfaces/user';
import validatePresence from '../validators/validate_presence';
import validEmail from '../validators/valid_email';

const baseClass = 'invite-user-form';

class InviteUserForm extends Component {
  static propTypes = {
    error: PropTypes.string,
    invitedBy: userInterface,
    onCancel: PropTypes.func,
    onSubmit: PropTypes.func,
  };

  constructor (props) {
    super(props);

    this.state = {
      errors: {
        admin: null,
        email: null,
        name: null,
      },
      formData: {
        admin: 'false',
        email: '',
        name: '',
      },
    };
  }

  componentWillReceiveProps (nextProps) {
    const { error } = nextProps;
    const { errors } = this.state;

    if (this.props.error !== error) {
      this.setState({
        errors: {
          ...errors,
          email: error,
        },
      });
    }
  }

  onInputChange = (formField) => {
    return (value) => {
      const { errors, formData } = this.state;

      this.setState({
        errors: {
          ...errors,
          [formField]: null,
        },
        formData: {
          ...formData,
          [formField]: value,
        },
      });

      return false;
    };
  }

  onRadioInputChange = (formField) => {
    return (evt) => {
      const { value } = evt.target;

      return this.onInputChange(formField)(value);
    };
  };

  onFormSubmit = (evt) => {
    evt.preventDefault();
    const valid = this.validate();

    if (valid) {
      const { formData: { admin, email, name } } = this.state;
      const { invitedBy, onSubmit } = this.props;

      return onSubmit({
        admin: admin === 'true',
        email,
        invited_by: invitedBy.id,
        name,
      });
    }

    return false;
  }

  validate = () => {
    const {
      errors,
      formData: { email },
    } = this.state;

    if (!validatePresence(email)) {
      this.setState({
        errors: {
          ...errors,
          email: 'Email field must be completed',
        },
      });

      return false;
    }

    if (!validEmail(email)) {
      this.setState({
        errors: {
          ...errors,
          email: `${email} is not a valid email`,
        },
      });

      return false;
    }

    return true;
  }

  render () {
    const { errors, formData: { admin, email, name } } = this.state;
    const { onCancel } = this.props;
    const { onFormSubmit, onInputChange, onRadioInputChange } = this;

    return (
      <form onSubmit={onFormSubmit}>
        <InputFieldWithIcon
          autofocus
          error={errors.name}
          name="name"
          onChange={onInputChange('name')}
          placeholder="Name"
          value={name}
        />
        <InputFieldWithIcon
          error={errors.email}
          name="email"
          onChange={onInputChange('email')}
          placeholder="Email"
          value={email}
        />
        <div className={`${baseClass}__radio`}>
          <p className={`${baseClass}__role`}>role</p>
          <input
            checked={admin === 'false'}
            onChange={onRadioInputChange('admin')}
            type="radio"
            value="false"
          /> USER (default)
          <br />
          <input
            checked={admin === 'true'}
            onChange={onRadioInputChange('admin')}
            type="radio"
            value="true"
          /> ADMIN
        </div>
        <div className={`${baseClass}__btn-wrap`}>
          <Button
            className={`${baseClass}__btn`}
            text="Invite"
            type="submit"
          />
          <Button
            className={`${baseClass}__btn`}
            onClick={onCancel}
            text="Cancel"
            type="input"
            variant="inverse"
          />
        </div>
      </form>
    );
  }
}

export default InviteUserForm;